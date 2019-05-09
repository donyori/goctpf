package prefab

import (
	"github.com/donyori/gocontainer"
	"github.com/donyori/gocontainer/pqueue"
	"github.com/donyori/goctpf"
)

type Comparable = gocontainer.Comparable

type PriorityTaskManager struct {
	pq *pqueue.PriorityQueue
}

const pqInitCap int = 64

func NewPriorityTaskManager() *PriorityTaskManager {
	pq := pqueue.NewPriorityQueue(pqInitCap, true, false)
	return &PriorityTaskManager{pq: pq}
}

func PriorityTaskManagerMaker() goctpf.TaskManager {
	return NewPriorityTaskManager()
}

func (ptm *PriorityTaskManager) Init() {
	if ptm.pq != nil {
		ptm.pq.Reset(pqInitCap)
	} else {
		ptm.pq = pqueue.NewPriorityQueue(pqInitCap, true, false)
	}
}

func (ptm *PriorityTaskManager) NumTask() int {
	if ptm == nil || ptm.pq == nil {
		return 0
	}
	return ptm.pq.Len()
}

func (ptm *PriorityTaskManager) Add(
	source goctpf.Source, tasks ...interface{}) error {
	if ptm.pq == nil {
		ptm.pq = pqueue.NewPriorityQueue(pqInitCap, true, false)
	}
	// Ensure type.
	xs := make([]Comparable, 0, len(tasks))
	for _, task := range tasks {
		xs = append(xs, task.(Comparable))
	}
	for _, x := range xs {
		ptm.pq.Enqueue(x)
	}
	return nil
}

func (ptm *PriorityTaskManager) Pick(purpose goctpf.Purpose) (
	task interface{}, err error) {
	if ptm == nil || ptm.pq == nil {
		return nil, goctpf.ErrNoMoreTask
	}
	task, ok := ptm.pq.Dequeue()
	if !ok {
		return nil, goctpf.ErrNoMoreTask
	}
	return
}

func (ptm *PriorityTaskManager) Peek(purpose goctpf.Purpose) (
	task interface{}, err error) {
	if ptm == nil || ptm.pq == nil || ptm.pq.Len() == 0 {
		return nil, goctpf.ErrNoMoreTask
	}
	task = ptm.pq.Top()
	return
}

func (ptm *PriorityTaskManager) Clear() {
	if ptm != nil && ptm.pq != nil {
		ptm.pq.Clear()
	}
}
