package cfw

import (
	"sync"
	"time"

	"github.com/donyori/goctpf"
	"github.com/donyori/goctpf/idtpf"
	"github.com/donyori/goctpf/internal/util"
	"github.com/donyori/gorecover"
)

func workerProc(workerNo int,
	taskHandler idtpf.TaskHandler,
	setup goctpf.Setup,
	tearDown goctpf.TearDown,
	sendErrTimeout time.Duration,
	runningWg, taskWg *sync.WaitGroup,
	taskInChan <-chan interface{},
	taskOutChan chan<- interface{},
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

	// Some variables used in the main loop:
	var task interface{}
	var newTasks []interface{}
	var newTasksLen, sentCount, unsentCount, i int
	var doesExit bool
	var errToPanic error
	errBuf := make([]error, 0, 4)
	doesContinue := true

	// The main loop:
	for doesContinue {
		select {
		case <-exitInChan: // Receive exit signal from main proc.
			doesContinue = false
		case <-doneChan: // Receive done signal from worker supervisor.
			doesContinue = false
		case task = <-taskInChan: // Receive task from main proc.
			func() {
				defer taskWg.Done() // Make sure taskWg.Done() can be executed at last.
				errBuf = errBuf[:0] // Clear errBuf, but keep the underlying array.
				errToPanic = gorecover.Recover(func() {
					defer util.DoneTask(task)
					newTasks, doesExit = taskHandler(workerNo, task, &errBuf)
					newTasksLen = len(newTasks)
					sentCount = 0
					defer func() {
						for i = sentCount; i < newTasksLen; i++ {
							util.DiscardTask(newTasks[i])
						}
					}()
					if doesExit {
						doesContinue = false
						return
					}
					if newTasksLen == 0 {
						return
					}
					taskWg.Add(newTasksLen)
					defer func() {
						unsentCount = newTasksLen - sentCount
						if unsentCount > 0 {
							// This case is the main proc asks to exit before sending new tasks finish.
							taskWg.Add(-unsentCount) // Adjust task counting.
						}
					}()
					for sentCount < newTasksLen {
						select {
						case <-exitInChan: // Receive exit signal from main proc.
							doesContinue = false
							return
						case taskOutChan <- newTasks[sentCount]:
							sentCount++
						}
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
			// End of "case task, ok = <-taskInChan" in main select.
		} // End of main select.
	} // End of main loop.
}
