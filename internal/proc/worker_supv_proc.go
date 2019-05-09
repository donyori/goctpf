package proc

import "sync"

func WorkerSupvProc(runningWg, taskWg *sync.WaitGroup,
	doneToWorkersChan, doneToMainChan chan<- struct{}) {
	defer close(doneToMainChan)
	// Wait for all tasks done.
	taskWg.Wait()
	// Broadcast done signal to workers.
	close(doneToWorkersChan)
	// Wait for all workers exiting.
	runningWg.Wait()
}
