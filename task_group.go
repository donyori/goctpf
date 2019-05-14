package goctpf

import (
	"errors"
	"sync"
)

type TaskGroup struct {
	wg              sync.WaitGroup
	doneH, discardH func(task interface{})
}

type TaskGroupMember struct {
	Task interface{}

	tg   *TaskGroup
	once sync.Once // For Done or Discard.
}

func NewTaskGroup(doneHandler, discardHandler func(task interface{})) *TaskGroup {
	return &TaskGroup{doneH: doneHandler, discardH: discardHandler}
}

func (tg *TaskGroup) WrapTask(task interface{}) *TaskGroupMember {
	if tg == nil {
		panic(errors.New("goctpf: task group is nil"))
	}
	tg.wg.Add(1)
	return &TaskGroupMember{Task: task, tg: tg}
}

func (tg *TaskGroup) Wait() {
	if tg == nil {
		return
	}
	tg.wg.Wait()
}

func (tgm *TaskGroupMember) GetTaskGroup() *TaskGroup {
	if tgm == nil {
		return nil
	}
	return tgm.tg
}

func (tgm *TaskGroupMember) Done() {
	if tgm == nil {
		return
	}
	tgm.once.Do(func() {
		tgm.tg.wg.Done()
		if tgm.tg.doneH != nil {
			tgm.tg.doneH(tgm)
		}
	})
}

func (tgm *TaskGroupMember) Discard() {
	if tgm == nil {
		return
	}
	tgm.once.Do(func() {
		tgm.tg.wg.Done()
		if tgm.tg.discardH != nil {
			tgm.tg.discardH(tgm)
		}
	})
}
