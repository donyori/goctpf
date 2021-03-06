package cfw

import (
	"errors"
	"runtime"
	"sync"

	"github.com/donyori/goctpf"
	"github.com/donyori/goctpf/idtpf"
	"github.com/donyori/goctpf/internal/proc"
	"github.com/donyori/goctpf/internal/util"
)

func mainProc(taskMgr goctpf.TaskManager,
	taskHandler idtpf.TaskHandler,
	setup goctpf.Setup,
	tearDown goctpf.TearDown,
	workerSettings goctpf.WorkerSettings,
	taskChan <-chan interface{},
	workerErrChan chan<- error) {
	if taskMgr == nil {
		panic(errors.New("goctpf: taskMgr is nil"))
	}
	if taskHandler == nil {
		panic(errors.New("goctpf: taskHandler is nil"))
	}

	numWorker := workerSettings.Number
	if numWorker == 0 {
		maxprocs := runtime.GOMAXPROCS(0)
		if maxprocs > 0 {
			numWorker = uint32(maxprocs)
		} else {
			numWorker = 1
		}
	}

	// Initialize task manager:
	taskMgr.Init()

	// Wait groups:
	// runningWg for check whether workers are exited or not.
	// taskWg for check whether all tasks are done or not.
	var runningWg, taskWg sync.WaitGroup

	// Channels:
	stChan := make(chan interface{})         // send task to workers
	rtChan := make(chan interface{})         // receive new task from workers. DON'T CLOSE IT!
	seChan := make(chan struct{})            // broadcast exit signal to workers
	reChan := make(chan struct{}, numWorker) // receive exit quest from workers
	dChan := make(chan struct{})             // receive done signal from worker supervisor
	dwChan := make(chan struct{})            // for worker supervisor to broadcast done signal to workers

	// Channels used in this goroutine:
	var taskOutChan chan<- interface{} // = nil, disable this channel at the beginning
	var taskInChan <-chan interface{} = rtChan
	var exitOutChan chan<- struct{} = seChan
	var exitInChan <-chan struct{} = reChan
	var doneChan <-chan struct{} = dChan

	curN := taskMgr.NumTask()
	var delta int // For adjusting task counting.
	// Defer to close exitOutChan and wait for workers and worker supervisor to exit.
	// Before workers starting, for safety.
	defer func() {
		close(exitOutChan)
		_ = taskMgr.Scan(util.DiscardTask) // Ignore error.
		// Remove unsent tasks, to avoid worker supervisor waiting forever.
		numUnsent := taskMgr.NumTask()
		if numUnsent-curN != delta { // Panic when add or pick task.
			taskWg.Add(numUnsent - curN - delta) // Adjust task counting.
		}
		if numUnsent > 0 {
			taskWg.Add(-numUnsent) // Adjust task counting.
		}
		taskMgr.Clear()
		// Done dummy task count if taskChan is enable, to avoid worker supervisor waiting forever.
		if taskChan != nil {
			taskWg.Done()
			var task interface{}
			// Drain taskChan and disable it for safety.
			for task = range taskChan {
				// Discard unsent tasks.
				util.DiscardTask(task)
			}
			taskChan = nil
		}
		// Wait for workers and worker supervisor exit.
		<-doneChan
	}()

	// Start workers and worker supervisor:
	taskWg.Add(1) // Add one dummy task count, standing for taskChan is open.
	for i, n := 0, int(numWorker); i < n; i++ {
		runningWg.Add(1)
		go workerProc(i, taskHandler, setup, tearDown,
			workerSettings.SendErrTimeout, &runningWg, &taskWg,
			stChan, rtChan, workerErrChan, dwChan, seChan, reChan)
	}
	go proc.WorkerSupvProc(&runningWg, &taskWg, dwChan, dChan)

	// Some variables used in the main loop:
	var task, nextToSendTask interface{}
	var ok bool
	var err error
	doesContinue := true

	// The main loop:
	for doesContinue {
		select {
		case <-exitInChan: // A worker asks to exit.
			doesContinue = false
		case task, ok = <-taskChan: // Receive a task from APP.
			if !ok {
				taskWg.Done()  // Done dummy task count, standing for taskChan is closed.
				taskChan = nil // Disable taskChan.
				break
			}
			delta = 1
			err = taskMgr.Add(goctpf.FromApp, task)
			if err != nil {
				// For Add(), discard task before panic.
				util.DiscardTask(task)
				break
			}
			curN = taskMgr.NumTask()
			delta = 0
			taskWg.Add(1)
			taskOutChan = stChan                                  // Enable taskOutChan.
			nextToSendTask, err = taskMgr.Peek(goctpf.ForWorkers) // Error will be checked after select.
		case task = <-taskInChan: // Receive a task from a worker.
			delta = 1
			err = taskMgr.Add(goctpf.FromWorkers, task)
			if err != nil {
				// For Add(), discard task before panic.
				util.DiscardTask(task)
				break
			}
			curN = taskMgr.NumTask()
			delta = 0
			// taskWg.Add() executed in worker.
			taskOutChan = stChan                                  // Enable taskOutChan.
			nextToSendTask, err = taskMgr.Peek(goctpf.ForWorkers) // Error will be checked after select.
		case taskOutChan <- nextToSendTask: // After sending a task to a worker.
			delta = -1
			_, err = taskMgr.Pick(goctpf.ForWorkers)
			if err != nil {
				break
			}
			curN = taskMgr.NumTask()
			delta = 0
			if curN > 0 {
				nextToSendTask, err = taskMgr.Peek(goctpf.ForWorkers) // Error will be checked after select.
			} else {
				taskOutChan = nil // Disable taskOutChan.
				nextToSendTask = nil
			}
		case <-doneChan: // Receive done signal from worker supervisor.
			doesContinue = false
		}
		if err != nil {
			panic(err)
		}
	}
}
