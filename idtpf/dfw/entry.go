package dfw

import (
	"github.com/donyori/goctpf"
	"github.com/donyori/goctpf/idtpf"
	"github.com/donyori/goctpf/internal/util"
	"github.com/donyori/goctpf/prefab"
)

func Start(taskHandler idtpf.TaskHandler,
	taskChan <-chan interface{},
	workerErrChan chan<- error) {
	go mainProc(prefab.LdgbTaskManagerMaker, taskHandler,
		*goctpf.NewWorkerSettings(), taskChan, workerErrChan)
}

func StartEx(taskMgrMaker goctpf.TaskManagerMaker,
	taskHandler idtpf.TaskHandler,
	workerSettings goctpf.WorkerSettings,
	taskChan <-chan interface{},
	workerErrChan chan<- error) {
	go mainProc(taskMgrMaker, taskHandler,
		workerSettings, taskChan, workerErrChan)
}

func Do(taskHandler idtpf.TaskHandler,
	workerErrChan chan<- error,
	initialTasks ...interface{}) {
	if len(initialTasks) == 0 {
		return
	}
	mainProc(prefab.LdgbTaskManagerMaker, taskHandler,
		*goctpf.NewWorkerSettings(), util.InitialTasksToChan(initialTasks...),
		workerErrChan)
}

func DoEx(taskMgrMaker goctpf.TaskManagerMaker,
	taskHandler idtpf.TaskHandler,
	workerSettings goctpf.WorkerSettings,
	workerErrChan chan<- error,
	initialTasks ...interface{}) {
	if len(initialTasks) == 0 {
		return
	}
	mainProc(taskMgrMaker, taskHandler, workerSettings,
		util.InitialTasksToChan(initialTasks...), workerErrChan)
}
