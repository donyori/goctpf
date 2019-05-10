package util

import "github.com/donyori/goctpf"

func PostProcessingOfTaskHandling(task interface{}) {
	if tgm, ok := task.(*goctpf.TaskGroupMember); ok {
		tg := tgm.GetTaskGroup()
		tg.Done()
	}
}
