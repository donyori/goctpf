package util

func InitialTasksToChan(initialTasks ...interface{}) <-chan interface{} {
	taskChan := make(chan interface{}, len(initialTasks))
	for _, task := range initialTasks {
		taskChan <- task
	}
	close(taskChan)
	return taskChan
}
