package goctpf

import (
	"errors"
	"sync"
	"sync/atomic"
)

type TaskGroup struct {
	doneChan   <-chan struct{}
	notifyChan chan struct{}
	counter    int64
}

type TaskGroupMember struct {
	Task interface{}

	tg       *TaskGroup
	doneOnce sync.Once
}

func NewTaskGroup(doneChan <-chan struct{}) *TaskGroup {
	return &TaskGroup{
		doneChan:   doneChan,
		notifyChan: make(chan struct{}, 1),
	}
}

func (tg *TaskGroup) WrapTask(task interface{}) *TaskGroupMember {
	if tg == nil {
		panic(errors.New("goctpf: task group is nil"))
	}
	atomic.AddInt64(&tg.counter, 1)
	return &TaskGroupMember{Task: task, tg: tg}
}

func (tg *TaskGroup) Wait() {
	if tg == nil {
		return
	}
	c := atomic.LoadInt64(&tg.counter)
	for c > 0 {
		select {
		case <-tg.notifyChan:
			c = atomic.LoadInt64(&tg.counter)
		case <-tg.doneChan:
			c = 0
		}
	}
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
	tgm.doneOnce.Do(func() {
		c := atomic.AddInt64(&tgm.tg.counter, -1)
		if c < 0 {
			panic(errors.New("goctpf: counter of TaskGroup is negative"))
		}
		tgm.tg.notifyChan <- struct{}{}
	})
}
