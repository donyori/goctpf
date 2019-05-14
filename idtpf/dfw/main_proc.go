package dfw

import (
	"errors"
	"runtime"
	"sync"

	"github.com/donyori/goctpf"
	"github.com/donyori/goctpf/idtpf"
	"github.com/donyori/goctpf/internal/proc"
	"github.com/donyori/goctpf/internal/util"
)

func mainProc(taskMgrMaker goctpf.TaskManagerMaker,
	taskHandler idtpf.TaskHandler,
	setup goctpf.Setup,
	tearDown goctpf.TearDown,
	workerSettings goctpf.WorkerSettings,
	appTaskChan <-chan interface{},
	workerErrChan chan<- error) {
	if taskMgrMaker == nil {
		panic(errors.New("goctpf: taskMgrMaker is nil"))
	}
	if taskHandler == nil {
		panic(errors.New("goctpf: taskHandler is nil"))
	}

	n := int(workerSettings.Number)
	if n <= 0 {
		maxprocs := runtime.GOMAXPROCS(0)
		if maxprocs > 0 {
			n = maxprocs
		} else {
			n = 1
		}
	}

	// Wait groups:
	// runningWg for check whether workers are exited or not.
	// taskWg for check whether all tasks are done or not.
	var runningWg, taskWg sync.WaitGroup

	// A sync.Once for Done dummy task count, standing for taskChan is closed.
	var doneDummyOnce sync.Once

	// Channels:
	seChan := make(chan struct{})      // broadcast exit signal to workers
	reChan := make(chan struct{}, n)   // receive exit quest from workers
	dChan := make(chan struct{})       // receive done signal from worker supervisor
	tChan := make(chan interface{}, 1) // for workers to send tasks to each other. DON'T CLOSE IT!
	dwChan := make(chan struct{})      // for worker supervisor to broadcast done signal to workers

	// Channels used in this goroutine:
	var exitOutChan chan<- struct{} = seChan
	var exitInChan <-chan struct{} = reChan
	var doneChan <-chan struct{} = dChan
	var taskChan <-chan interface{} = tChan

	// Defer to close out channels and wait for workers and worker supervisor to exit.
	// Before workers starting, for safety.
	defer func() {
		close(exitOutChan)
		var task interface{}
		// Done dummy task count if no one did it, to avoid worker supervisor waiting forever.
		doneDummyOnce.Do(func() {
			taskWg.Done()
			// Drain appTaskChan and discard undone tasks.
			for task = range appTaskChan {
				util.DiscardTask(task)
			}
		})
		// Drain tChan and adjust task counting, to avoid worker supervisor waiting forever.
		for {
			select {
			case <-doneChan:
				// Workers and worker supervisor exited. Return now.
				// taskChan must be empty now.
				return
			case task = <-taskChan:
				taskWg.Done()
				// Discard the task.
				util.DiscardTask(task)
			}
		}
	}()

	// Start workers and worker supervisor:
	taskWg.Add(1) // Add one dummy task count, standing for taskChan is open.
	for i := 0; i < n; i++ {
		runningWg.Add(1)
		go workerProc(i, taskMgrMaker, taskHandler, setup, tearDown,
			workerSettings.SendErrTimeout, &runningWg, &taskWg, &doneDummyOnce,
			appTaskChan, tChan, workerErrChan, dwChan, seChan, reChan)
	}
	go proc.WorkerSupvProc(&runningWg, &taskWg, dwChan, dChan)

	// Wait for a worker asking to exit or all tasks done:
	select {
	case <-exitInChan: // A worker asks to exit.
	case <-doneChan: // Receive done signal from worker supervisor.
	}
}
