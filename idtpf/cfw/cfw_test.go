package cfw

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/donyori/goctpf"
	"github.com/donyori/goctpf/idtpf"
)

func TestCFw(t *testing.T) {
	initialTasks := []interface{}{0, 1, 3, 5}
	errChan := make(chan error, 10)
	doneChan := make(chan struct{})
	defer func() {
		close(errChan)
		<-doneChan
	}()
	go func() {
		defer close(doneChan)
		for err := range errChan {
			fmt.Println(err)
			if err != errTestError {
				t.Error(err)
			}
		}
	}()
	handlers := []idtpf.TaskHandler{
		testTaskHandler1,
		testTaskHandler2,
		testTaskHandlerForTaskGroup1,
		testTaskHandlerForTaskGroup2,
	}
	for i := range handlers {
		if i < 2 {
			t.Run("handler"+strconv.Itoa(i+1)+"-Do", func(t *testing.T) {
				Do(handlers[i], nil, nil, errChan, initialTasks...)
			})
			t.Run("handler"+strconv.Itoa(i+1)+"-DoEx", func(t *testing.T) {
				DoEx(new(testTaskMgr),
					handlers[i],
					nil,
					nil,
					goctpf.WorkerSettings{Number: 4, SendErrTimeout: time.Nanosecond},
					errChan,
					initialTasks...)
			})
		} else {
			t.Run("handlerForTaskGroup"+strconv.Itoa(i-1)+"-Start", func(t *testing.T) {
				taskChan := make(chan interface{}, len(initialTasks))
				dc := Start(handlers[i], nil, nil, taskChan, errChan)
				testTaskGroup = goctpf.NewTaskGroup(nil, nil)
				for i := range initialTasks {
					tgm := testTaskGroup.WrapTask(initialTasks[i])
					taskChan <- tgm
				}
				close(taskChan)
				testTaskGroup.Wait()
				// Wait multiple times:
				testTaskGroup.Wait()
				testTaskGroup.Wait()
				<-dc
			})
			t.Run("handlerForTaskGroup"+strconv.Itoa(i-1)+"-StartEx", func(t *testing.T) {
				taskChan := make(chan interface{}, len(initialTasks))
				dc := StartEx(new(testTaskMgr),
					handlers[i],
					nil,
					nil,
					goctpf.WorkerSettings{Number: 4, SendErrTimeout: time.Nanosecond},
					taskChan,
					errChan)
				testTaskGroup = goctpf.NewTaskGroup(nil, nil)
				for i := range initialTasks {
					tgm := testTaskGroup.WrapTask(initialTasks[i])
					taskChan <- tgm
				}
				close(taskChan)
				testTaskGroup.Wait()
				// Wait multiple times:
				testTaskGroup.Wait()
				testTaskGroup.Wait()
				<-dc
			})
		}
	}
}
