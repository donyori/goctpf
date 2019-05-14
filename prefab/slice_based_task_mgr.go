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

func (sbtm *sliceBasedTaskManager) Scan(
	handler func(task interface{}) (doesStop bool)) error {
	if sbtm == nil || handler == nil {
		return nil
	}
	for i := len(sbtm.a) - 1; i >= 0; i-- {
		if handler(sbtm.a[i]) {
			return nil
		}
	}
	return nil
}

func (sbtm *sliceBasedTaskManager) Clear() {
	if sbtm == nil {
		return
	}
	// The same as Init().
	sbtm.Init()
}
