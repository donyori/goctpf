package prefab

import "github.com/donyori/goctpf"

// Depth-first locally, breadth-first globally.
type LdgbTaskManager struct {
	sliceBasedTaskManager
}

func NewLdgbTaskManager() *LdgbTaskManager {
	return new(LdgbTaskManager)
}

func LdgbTaskManagerMaker() goctpf.TaskManager {
	return NewLdgbTaskManager()
}

func (ldgbtm *LdgbTaskManager) Add(
	source goctpf.Source, tasks ...interface{}) error {
	switch source {
	case goctpf.FromApp, goctpf.FromOthers:
		if len(tasks) == 0 {
			return nil
		}
		// Add tasks to front.
		ldgbtm.a = append(ldgbtm.a, tasks...)
		copy(ldgbtm.a[len(tasks):], ldgbtm.a)
		copy(ldgbtm.a, tasks)
	case goctpf.FromWorkers, goctpf.FromMe:
		if len(tasks) == 0 {
			return nil
		}
		// Add tasks to back.
		l := len(ldgbtm.a)
		ldgbtm.a = append(ldgbtm.a, tasks...)
		// Reverse new items in order to let the more prior tasks be executed earlier.
		for r := len(ldgbtm.a) - 1; l < r; l, r = l+1, r-1 {
			ldgbtm.a[l], ldgbtm.a[r] = ldgbtm.a[r], ldgbtm.a[l]
		}
	default:
		return goctpf.NewUnknownSourceError(source)
	}
	return nil
}

func (ldgbtm *LdgbTaskManager) Pick(purpose goctpf.Purpose) (
	task interface{}, err error) {
	if ldgbtm == nil || len(ldgbtm.a) == 0 {
		return nil, goctpf.ErrNoMoreTask
	}
	last := len(ldgbtm.a) - 1
	old := ldgbtm.a
	switch purpose {
	case goctpf.ForWorkers, goctpf.ForMe:
		task = old[last]
		old[last] = nil // To avoid potential memory leak.
		ldgbtm.a = old[:last]
	case goctpf.ForOthers:
		task = old[0]
		old[0] = nil // To avoid potential memory leak.
		ldgbtm.a = old[1:]
	default:
		return nil, goctpf.NewUnknownPurposeError(purpose)
	}
	return task, nil
}

func (ldgbtm *LdgbTaskManager) Peek(purpose goctpf.Purpose) (
	task interface{}, err error) {
	if ldgbtm == nil || len(ldgbtm.a) == 0 {
		return nil, goctpf.ErrNoMoreTask
	}
	switch purpose {
	case goctpf.ForWorkers, goctpf.ForMe:
		task = ldgbtm.a[len(ldgbtm.a)-1]
	case goctpf.ForOthers:
		task = ldgbtm.a[0]
	default:
		return nil, goctpf.NewUnknownPurposeError(purpose)
	}
	return task, nil
}
