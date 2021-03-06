package dfw

import (
	"github.com/donyori/goctpf"
	"github.com/donyori/goctpf/idtpf"
	"github.com/donyori/goctpf/internal/util"
	"github.com/donyori/goctpf/prefab"
)

func Start(taskHandler idtpf.TaskHandler,
	setup goctpf.Setup,
	tearDown goctpf.TearDown,
	taskChan <-chan interface{},
	workerErrChan chan<- error) <-chan struct{} {
	doneChan := make(chan struct{})
	go func() {
		defer close(doneChan)
		mainProc(prefab.LdgbTaskManagerMaker, taskHandler, setup, tearDown,
			*goctpf.NewWorkerSettings(), taskChan, workerErrChan)
	}()
	return doneChan
}

func StartEx(taskMgrMaker goctpf.TaskManagerMaker,
	taskHandler idtpf.TaskHandler,
	setup goctpf.Setup,
	tearDown goctpf.TearDown,
	workerSettings goctpf.WorkerSettings,
	taskChan <-chan interface{},
	workerErrChan chan<- error) <-chan struct{} {
	doneChan := make(chan struct{})
	go func() {
		defer close(doneChan)
		mainProc(taskMgrMaker, taskHandler, setup, tearDown,
			workerSettings, taskChan, workerErrChan)
	}()
	return doneChan
}

func Do(taskHandler idtpf.TaskHandler,
	setup goctpf.Setup,
	tearDown goctpf.TearDown,
	workerErrChan chan<- error,
	initialTasks ...interface{}) {
	if len(initialTasks) == 0 {
		return
	}
	mainProc(prefab.LdgbTaskManagerMaker, taskHandler, setup, tearDown,
		*goctpf.NewWorkerSettings(), util.InitialTasksToChan(initialTasks...),
		workerErrChan)
}

func DoEx(taskMgrMaker goctpf.TaskManagerMaker,
	taskHandler idtpf.TaskHandler,
	setup goctpf.Setup,
	tearDown goctpf.TearDown,
	workerSettings goctpf.WorkerSettings,
	workerErrChan chan<- error,
	initialTasks ...interface{}) {
	if len(initialTasks) == 0 {
		return
	}
	mainProc(taskMgrMaker, taskHandler, setup, tearDown, workerSettings,
		util.InitialTasksToChan(initialTasks...), workerErrChan)
}
