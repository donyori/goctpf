package dfw

import (
	"errors"
	"sync"

	"github.com/donyori/goctpf"
	"github.com/donyori/goctpf/idtpf"
	"github.com/donyori/goctpf/internal/proc"
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
	if workerSettings.Number == 0 {
		panic(errors.New("goctpf: the number of workers is 0"))
	}

	n := int(workerSettings.Number)

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
		// Done dummy task count if no one did it, to avoid worker supervisor waiting forever.
		doneDummyOnce.Do(func() {
			taskWg.Done()
			// Drain appTaskChan for safety.
			for range appTaskChan {
			}
		})
		// Drain tChan and adjust task counting, to avoid worker supervisor waiting forever.
		for {
			select {
			case <-doneChan:
				// Workers and worker supervisor exited. Return now.
				return
			case <-taskChan:
				taskWg.Done()
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
