package util

import "github.com/donyori/goctpf"

func DoneTask(task interface{}) bool {
	if task == nil {
		return false
	}
	if t, ok := task.(goctpf.DoneProcessor); ok {
		t.Done()
	}
	return false
}

func DiscardTask(task interface{}) bool {
	if task == nil {
		return false
	}
	if t, ok := task.(goctpf.DiscardProcessor); ok {
		t.Discard()
	}
	return false
}
