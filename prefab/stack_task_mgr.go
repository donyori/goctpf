package prefab

import "github.com/donyori/goctpf"

type StackTaskManager struct {
	sliceBasedTaskManager
}

func NewStackTaskManager() *StackTaskManager {
	return new(StackTaskManager)
}

func StackTaskManagerMaker() goctpf.TaskManager {
	return NewStackTaskManager()
}

func (stm *StackTaskManager) Add(
	source goctpf.Source, tasks ...interface{}) error {
	stm.a = append(stm.a, tasks...)
	return nil
}

func (stm *StackTaskManager) Pick(purpose goctpf.Purpose) (
	task interface{}, err error) {
	if stm == nil || len(stm.a) == 0 {
		return nil, goctpf.ErrNoMoreTask
	}
	last := len(stm.a) - 1
	old := stm.a
	task = old[last]
	old[last] = nil // To avoid potential memory leak.
	stm.a = old[:last]
	return task, nil
}

func (stm *StackTaskManager) Peek(purpose goctpf.Purpose) (
	task interface{}, err error) {
	if stm == nil || len(stm.a) == 0 {
		return nil, goctpf.ErrNoMoreTask
	}
	return stm.a[len(stm.a)-1], nil
}
