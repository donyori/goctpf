package util

import "time"

// Ensure errChan != nil && len(errs) > 0.
func SendErrors(timer *time.Timer, sendErrTimeout time.Duration,
	errChan chan<- error, exitChan <-chan struct{}, errs ...error) (
	doesContinue bool) {
	doesContinue = true
	isTimeout := false
	var timeoutChan <-chan time.Time
	if timer != nil {
		timeoutChan = timer.C
		ResetTimer(timer, sendErrTimeout)
		defer timer.Stop()
	}
	for _, err := range errs {
		if isTimeout {
			// Try to send error immediately.
			select {
			case errChan <- err:
			default:
				return
			}
		} else {
			select {
			case errChan <- err: // !Before check exit and done signal! To send as many errors as possible.
				// Do nothing else.
			case <-exitChan: // Receive exit signal from main proc.
				doesContinue = false
				isTimeout = true
			case <-timeoutChan:
				isTimeout = true
			}
		} // End of "if isTimeout".
	} // End of "for _, err = range errBuf".
	return
}
