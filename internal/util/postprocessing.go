package util

import "github.com/donyori/goctpf"

func DoneTask(task interface{}) bool {
	if task == nil {
		return false
	}
	if tgm, ok := task.(*goctpf.TaskGroupMember); ok {
		tgm.Done()
	}
	return false
}

func DiscardTask(task interface{}) bool {
	if task == nil {
		return false
	}
	if tgm, ok := task.(*goctpf.TaskGroupMember); ok {
		tgm.Discard()
	}
	return false
}
