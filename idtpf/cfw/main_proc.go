package cfw

import (
	"errors"
	"sync"

	"github.com/donyori/goctpf"
	"github.com/donyori/goctpf/idtpf"
	"github.com/donyori/goctpf/internal/proc"
)

func mainProc(taskMgr goctpf.TaskManager,
	taskHandler idtpf.TaskHandler,
	setupAndTearDown *goctpf.SetupAndTearDown,
	workerSettings goctpf.WorkerSettings,
	taskChan <-chan interface{},
	workerErrChan chan<- error) {
	if taskMgr == nil {
		panic(errors.New("goctpf: taskMgr is nil"))
	}
	if taskHandler == nil {
		panic(errors.New("goctpf: taskHandler is nil"))
	}
	if workerSettings.Number == 0 {
		panic(errors.New("goctpf: the number of workers is 0"))
	}

	// Initialize task manager:
	taskMgr.Init()

	// Wait groups:
	// runningWg for check whether workers are exited or not.
	// taskWg for check whether all tasks are done or not.
	var runningWg, taskWg sync.WaitGroup

	// Channels:
	stChan := make(chan interface{})                     // send task to workers
	rtChan := make(chan interface{})                     // receive new task from workers. DON'T CLOSE IT!
	seChan := make(chan struct{})                        // broadcast exit signal to workers
	reChan := make(chan struct{}, workerSettings.Number) // receive exit quest from workers
	dChan := make(chan struct{})                         // receive done signal from worker supervisor
	dwChan := make(chan struct{})                        // for worker supervisor to broadcast done signal to workers

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
			// Drain taskChan and disable it for safety.
			for range taskChan {
			}
			taskChan = nil
		}
		// Wait for workers and worker supervisor exit.
		<-doneChan
	}()

	// Start workers and worker supervisor:
	taskWg.Add(1) // Add one dummy task count, standing for taskChan is open.
	for i, n := 0, int(workerSettings.Number); i < n; i++ {
		runningWg.Add(1)
		go workerProc(i, taskHandler, setupAndTearDown,
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
