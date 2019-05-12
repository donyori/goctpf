package idtpf

type TaskHandler func(workerNo int, task interface{}, errBuf *[]error) (
	newTasks []interface{}, doesExit bool)
