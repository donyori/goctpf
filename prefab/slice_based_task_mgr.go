package prefab

type sliceBasedTaskManager struct {
	a []interface{}
}

func (sbtm *sliceBasedTaskManager) Init() {
	sbtm.a = nil
}

func (sbtm *sliceBasedTaskManager) NumTask() int {
	if sbtm == nil {
		return 0
	}
	return len(sbtm.a)
}

func (sbtm *sliceBasedTaskManager) Clear() {
	if sbtm == nil {
		return
	}
	// The same as Init().
	sbtm.Init()
}
