package idtpf

type TaskHandler func(task interface{}, errBuf *[]error) (
	newTasks []interface{}, doesExit bool)
