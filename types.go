package goctpf

type TaskManager interface {
	// Initialize task manager.
	Init()

	// Return the number of tasks.
	// This method should return 0 when TaskManager is nil.
	NumTask() int

	// Add tasks.
	Add(source Source, tasks ...interface{}) error

	// Pick out a task according to the purpose.
	// Should return goctpf.ErrNoMoreTask if there is no task.
	Pick(purpose Purpose) (task interface{}, err error)

	// Peek a task according to the purpose.
	// Should return goctpf.ErrNoMoreTask if there is no task.
	Peek(purpose Purpose) (task interface{}, err error)

	// Clear task manager.
	// Sometimes it does the same thing as Init().
	// This method should return directly when TaskManager is nil.
	Clear()
}

type TaskManagerMaker func() TaskManager
