package cfw

import (
	"errors"
	"fmt"

	"github.com/donyori/goctpf"
)

type testTaskMgr []interface{}

var errTestError error = errors.New("error for test")
var testTaskGroup *goctpf.TaskGroup

func testTaskHandler1(task interface{}, errBuf *[]error) (
	newTasks []interface{}, doesExit bool) {
	fmt.Println(task)
	x := task.(int)
	if x < 10 {
		newTasks = append(newTasks, x+1)
	}
	if x == 4 {
		*errBuf = append(*errBuf, errTestError)
	}
	if x == 7 {
		newTasks = append(newTasks, x+2)
	}
	return
}

func testTaskHandler2(task interface{}, errBuf *[]error) (
	newTasks []interface{}, doesExit bool) {
	fmt.Println(task)
	x := task.(int)
	for i := 0; i < 10; i++ {
		newTasks = append(newTasks, x+10)
	}
	if x > 30 {
		doesExit = true
	}
	return
}

func testTaskHandlerForTaskGroup1(task interface{}, errBuf *[]error) (
	newTasks []interface{}, doesExit bool) {
	t := task.(*goctpf.TaskGroupMember)
	fmt.Println(t.Task)
	x := t.Task.(int)
	if x < 10 {
		newTasks = append(newTasks, testTaskGroup.WrapTask(x+1))
	}
	return
}

func testTaskHandlerForTaskGroup2(task interface{}, errBuf *[]error) (
	newTasks []interface{}, doesExit bool) {
	t := task.(*goctpf.TaskGroupMember)
	fmt.Println(t.Task)
	x := t.Task.(int)
	for i := 0; i < 10; i++ {
		newTasks = append(newTasks, testTaskGroup.WrapTask(x+10))
	}
	if x > 30 {
		doesExit = true
	}
	return
}

func (ttm *testTaskMgr) Init() {
	*ttm = nil
}

func (ttm testTaskMgr) HasTask() bool {
	return len(ttm) > 0
}

func (ttm testTaskMgr) NumTask() int {
	return len(ttm)
}

func (ttm *testTaskMgr) Add(source goctpf.Source, tasks ...interface{}) error {
	*ttm = append(*ttm, tasks...)
	return nil
}

func (ttm *testTaskMgr) Pick(purpose goctpf.Purpose) (task interface{}, err error) {
	if len(*ttm) == 0 {
		return nil, goctpf.ErrNoMoreTask
	}
	old := *ttm
	last := len(old) - 1
	task = old[last]
	*ttm = old[:last]
	return
}

func (ttm testTaskMgr) Peek(purpose goctpf.Purpose) (task interface{}, err error) {
	if len(ttm) == 0 {
		return nil, goctpf.ErrNoMoreTask
	}
	return ttm[len(ttm)-1], nil
}

func (ttm *testTaskMgr) Clear() {
	ttm.Init()
}
