package prefab

import (
	"container/list"

	"github.com/donyori/goctpf"
)

type QueueTaskManager struct {
	ls *list.List
}

func NewQueueTaskManager() *QueueTaskManager {
	return &QueueTaskManager{ls: list.New()}
}

func QueueTaskManagerMaker() goctpf.TaskManager {
	return NewQueueTaskManager()
}

func (qtm *QueueTaskManager) Init() {
	if qtm.ls == nil {
		qtm.ls = list.New()
	} else {
		qtm.ls.Init()
	}
}

func (qtm *QueueTaskManager) NumTask() int {
	if qtm == nil || qtm.ls == nil {
		return 0
	}
	return qtm.ls.Len()
}

func (qtm *QueueTaskManager) Add(
	source goctpf.Source, tasks ...interface{}) error {
	if qtm.ls == nil {
		qtm.ls = list.New()
	}
	for _, task := range tasks {
		qtm.ls.PushBack(task)
	}
	return nil
}

func (qtm *QueueTaskManager) Pick(purpose goctpf.Purpose) (
	task interface{}, err error) {
	if qtm == nil || qtm.ls == nil {
		return nil, goctpf.ErrNoMoreTask
	}
	front := qtm.ls.Front()
	if front == nil {
		return nil, goctpf.ErrNoMoreTask
	}
	task = qtm.ls.Remove(front)
	return task, nil
}

func (qtm *QueueTaskManager) Peek(purpose goctpf.Purpose) (
	task interface{}, err error) {
	if qtm == nil || qtm.ls == nil {
		return nil, goctpf.ErrNoMoreTask
	}
	front := qtm.ls.Front()
	if front == nil {
		return nil, goctpf.ErrNoMoreTask
	}
	return front.Value, nil
}

func (qtm *QueueTaskManager) Clear() {
	if qtm != nil && qtm.ls != nil {
		qtm.ls.Init()
	}
}
