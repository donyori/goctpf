package cfw

import (
	"github.com/donyori/goctpf"
	"github.com/donyori/goctpf/idtpf"
	"github.com/donyori/goctpf/internal/util"
	"github.com/donyori/goctpf/prefab"
)

func Start(taskHandler idtpf.TaskHandler,
	taskChan <-chan interface{},
	workerErrChan chan<- error) <-chan struct{} {
	doneChan := make(chan struct{})
	go func() {
		defer close(doneChan)
		mainProc(prefab.NewLdgbTaskManager(), taskHandler,
			*goctpf.NewWorkerSettings(), taskChan, workerErrChan)
	}()
	return doneChan
}

func StartEx(taskMgr goctpf.TaskManager,
	taskHandler idtpf.TaskHandler,
	workerSettings goctpf.WorkerSettings,
	taskChan <-chan interface{},
	workerErrChan chan<- error) <-chan struct{} {
	doneChan := make(chan struct{})
	go func() {
		defer close(doneChan)
		mainProc(taskMgr, taskHandler, workerSettings, taskChan, workerErrChan)
	}()
	return doneChan
}

func Do(taskHandler idtpf.TaskHandler,
	workerErrChan chan<- error,
	initialTasks ...interface{}) {
	if len(initialTasks) == 0 {
		return
	}
	mainProc(prefab.NewLdgbTaskManager(), taskHandler,
		*goctpf.NewWorkerSettings(), util.InitialTasksToChan(initialTasks...),
		workerErrChan)
}

func DoEx(taskMgr goctpf.TaskManager,
	taskHandler idtpf.TaskHandler,
	workerSettings goctpf.WorkerSettings,
	workerErrChan chan<- error,
	initialTasks ...interface{}) {
	if len(initialTasks) == 0 {
		return
	}
	mainProc(taskMgr, taskHandler, workerSettings,
		util.InitialTasksToChan(initialTasks...), workerErrChan)
}
