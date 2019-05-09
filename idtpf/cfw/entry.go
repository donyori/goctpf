package cfw

import (
	"github.com/donyori/goctpf"
	"github.com/donyori/goctpf/idtpf"
	"github.com/donyori/goctpf/internal/util"
	"github.com/donyori/goctpf/prefab"
)

func Start(taskHandler idtpf.TaskHandler,
	taskChan <-chan interface{},
	workerErrChan chan<- error) {
	go mainProc(prefab.NewLdgbTaskManager(), taskHandler,
		*goctpf.NewWorkerSettings(), taskChan, workerErrChan)
}

func StartEx(taskMgr goctpf.TaskManager,
	taskHandler idtpf.TaskHandler,
	workerSettings goctpf.WorkerSettings,
	taskChan <-chan interface{},
	workerErrChan chan<- error) {
	go mainProc(taskMgr, taskHandler, workerSettings, taskChan, workerErrChan)
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
