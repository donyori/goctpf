package dfw

import (
	"sync"
	"time"

	"github.com/donyori/goctpf"
	"github.com/donyori/goctpf/idtpf"
	"github.com/donyori/goctpf/internal/util"
	"github.com/donyori/gorecover"
)

// This func will panic if the taskMgr return an error except goctpf.ErrNoMoreTask.
func workerProc(workerNo int,
	taskMgrMaker goctpf.TaskManagerMaker,
	taskHandler idtpf.TaskHandler,
	setup goctpf.Setup,
	tearDown goctpf.TearDown,
	sendErrTimeout time.Duration,
	runningWg, taskWg *sync.WaitGroup,
	doneDummyOnce *sync.Once,
	appTaskChan <-chan interface{},
	taskChan chan interface{},
	errChan chan<- error,
	doneChan, exitInChan <-chan struct{},
	exitOutChan chan<- struct{}) {
	defer runningWg.Done()
	defer func() {
		exitOutChan <- struct{}{}
	}()
	if setup != nil {
		setup(workerNo)
	}
	if tearDown != nil {
		defer tearDown(workerNo)
	}

	// Create a timer for sending error, if necessary:
	var timer *time.Timer
	if sendErrTimeout > 0 {
		timer = time.NewTimer(sendErrTimeout)
		// Just create a timer. Stop the timer now.
		timer.Stop()
	}

	// A func to handle error.
	panicErr := func(err error) {
		if errChan != nil {
			util.SendErrors(timer, sendErrTimeout, errChan, exitInChan, err)
		}
		panic(err)
	}

	// Create a task manager and initialize it:
	taskMgr := taskMgrMaker()
	taskMgr.Init()

	curN := taskMgr.NumTask()
	var delta int // For adjusting task counting.
	defer func() {
		// Remove undone tasks, to avoid worker supervisor waiting forever.
		numUndone := taskMgr.NumTask()
		if numUndone-curN != delta { // Panic when add or pick task.
			taskWg.Add(numUndone - curN - delta) // Adjust task counting.
		}
		if numUndone > 0 {
			taskWg.Add(-numUndone) // Adjust task counting.
		}
		taskMgr.Clear()
	}()

	// Some variables used in the main loop:
	var task, toSendTask interface{}
	var newTasks []interface{}
	var canSend, ok, doesExit bool
	errBuf := make([]error, 0, 4)
	var err, errToPanic error
	doesContinue := true

	// The main loop:
	for doesContinue {
		// Step 1:
		// Try to send undone tasks to other workers if the number of undone tasks is greater than 1 (i.e. at least 2):
		canSend = true
		for canSend && toSendTask != nil && curN > 1 {
			select {
			case <-exitInChan: // Receive exit signal from main proc.
				doesContinue = false
				canSend = false
			case <-doneChan: // Receive done signal from worker supervisor.
				doesContinue = false
				canSend = false
			case taskChan <- toSendTask:
				delta = -1
				_, err = taskMgr.Pick(goctpf.ForOthers)
				if err != nil {
					panicErr(err) // err CANNOT be goctpf.ErrNoMoreTask.
				}
				curN = taskMgr.NumTask()
				delta = 0
				toSendTask, err = taskMgr.Peek(goctpf.ForOthers)
				if err == goctpf.ErrNoMoreTask {
					toSendTask = nil
				} else if err != nil {
					panicErr(err)
				}
			default:
				canSend = false
			}
		}
		if !doesContinue {
			break
		}

		// Step 2:
		// Receive a task from other workers if the number of undone tasks is 0:
		for doesContinue && curN == 0 {
			select {
			case <-exitInChan: // Receive exit signal from main proc.
				doesContinue = false
			case <-doneChan: // Receive done signal from worker supervisor.
				doesContinue = false
			case task, ok = <-appTaskChan: // Receive a task from APP.
				if !ok {
					doneDummyOnce.Do(func() {
						taskWg.Done() // Done dummy task count, standing for taskChan is closed.
					})
					appTaskChan = nil // Disable appTaskChan.
					break
				}
				delta = 1
				err = taskMgr.Add(goctpf.FromApp, task)
				if err != nil {
					panicErr(err)
				}
				curN = taskMgr.NumTask()
				delta = 0
				taskWg.Add(1)
			case task = <-taskChan: // Receive a task from other workers.
				delta = 1
				err = taskMgr.Add(goctpf.FromOthers, task)
				if err != nil {
					panicErr(err)
				}
				curN = taskMgr.NumTask()
				delta = 0
				// taskWg.Add() executed in Step 3.
			}
		}
		if !doesContinue {
			break
		}

		// Step 3:
		// Pick and handle a task:
		func() {
			defer taskWg.Done() // Make sure taskWg.Done() can be executed at last.
			delta = -1
			task, err = taskMgr.Pick(goctpf.ForMe)
			if err != nil {
				panicErr(err) // err CANNOT be goctpf.ErrNoMoreTask.
			}
			curN = taskMgr.NumTask()
			delta = 0
			toSendTask, err = taskMgr.Peek(goctpf.ForOthers)
			if err == goctpf.ErrNoMoreTask {
				toSendTask = nil
			} else if err != nil {
				panicErr(err)
			}
			errBuf = errBuf[:0] // Clear errBuf, but keep the underlying array.
			errToPanic = gorecover.Recover(func() {
				defer util.PostProcessingOfTaskHandling(task)
				newTasks, doesExit = taskHandler(workerNo, task, &errBuf)
				if doesExit {
					doesContinue = false
					return
				}
				delta = len(newTasks)
				if delta > 0 {
					taskWg.Add(delta) // Adjust task counting before taskMgr.Add()!
				}
				err = taskMgr.Add(goctpf.FromMe, newTasks...)
				if err != nil {
					panic(err)
				}
				curN = taskMgr.NumTask()
				delta = 0
				toSendTask, err = taskMgr.Peek(goctpf.ForOthers)
				if err == goctpf.ErrNoMoreTask {
					toSendTask = nil
				} else if err != nil {
					panic(err)
				}
			}) // End of "gorecover.Recover()".
			if errToPanic != nil {
				errBuf = append(errBuf, errToPanic)
			}
			if errChan == nil || len(errBuf) == 0 {
				return
			}
			doesContinue = util.SendErrors(timer, sendErrTimeout,
				errChan, exitInChan, errBuf...) && doesContinue
		}()
		if errToPanic != nil {
			panic(errToPanic)
		}
		// End of Step 3.
	} // End of main loop.
}
