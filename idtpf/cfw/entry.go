package cfw

import (
	"github.com/donyori/goctpf"
	"github.com/donyori/goctpf/idtpf"
	"github.com/donyori/goctpf/internal/util"
	"github.com/donyori/goctpf/prefab"
)

func Start(taskHandler idtpf.TaskHandler,
	setupAndTearDown *goctpf.SetupAndTearDown,
	taskChan <-chan interface{},
	workerErrChan chan<- error) <-chan struct{} {
	doneChan := make(chan struct{})
	go func() {
		defer close(doneChan)
		mainProc(prefab.NewLdgbTaskManager(), taskHandler, setupAndTearDown,
			*goctpf.NewWorkerSettings(), taskChan, workerErrChan)
	}()
	return doneChan
}

func StartEx(taskMgr goctpf.TaskManager,
	taskHandler idtpf.TaskHandler,
	setupAndTearDown *goctpf.SetupAndTearDown,
	workerSettings goctpf.WorkerSettings,
	taskChan <-chan interface{},
	workerErrChan chan<- error) <-chan struct{} {
	doneChan := make(chan struct{})
	go func() {
		defer close(doneChan)
		mainProc(taskMgr, taskHandler, setupAndTearDown,
			workerSettings, taskChan, workerErrChan)
	}()
	return doneChan
}

func Do(taskHandler idtpf.TaskHandler,
	setupAndTearDown *goctpf.SetupAndTearDown,
	workerErrChan chan<- error,
	initialTasks ...interface{}) {
	if len(initialTasks) == 0 {
		return
	}
	mainProc(prefab.NewLdgbTaskManager(), taskHandler, setupAndTearDown,
		*goctpf.NewWorkerSettings(), util.InitialTasksToChan(initialTasks...),
		workerErrChan)
}

func DoEx(taskMgr goctpf.TaskManager,
	taskHandler idtpf.TaskHandler,
	setupAndTearDown *goctpf.SetupAndTearDown,
	workerSettings goctpf.WorkerSettings,
	workerErrChan chan<- error,
	initialTasks ...interface{}) {
	if len(initialTasks) == 0 {
		return
	}
	mainProc(taskMgr, taskHandler, setupAndTearDown, workerSettings,
		util.InitialTasksToChan(initialTasks...), workerErrChan)
}
