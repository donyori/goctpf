package dfw

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/donyori/goctpf"
	"github.com/donyori/goctpf/idtpf"
)

func TestDFw(t *testing.T) {
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
	handlers := []idtpf.TaskHandler{testTaskHandler1, testTaskHandler2}
	for i := range handlers {
		t.Run("handler"+strconv.Itoa(i+1)+"-Do", func(t *testing.T) {
			Do(handlers[i], errChan, initialTasks...)
		})
		t.Run("handler"+strconv.Itoa(i+1)+"-DoEx", func(t *testing.T) {
			DoEx(testTaskMgrMaker,
				handlers[i],
				goctpf.WorkerSettings{Number: 4, SendErrTimeout: time.Nanosecond},
				errChan,
				initialTasks...)
		})
	}
}
