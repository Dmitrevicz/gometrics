package agent

// Semaphore is a custom syncronization primitive.
type Semaphore struct {
	semaCh chan struct{}
}

// NewSemaphore creates Semaphore with buffered channel of size maxReq.
func NewSemaphore(maxReq int) *Semaphore {
	return &Semaphore{
		semaCh: make(chan struct{}, maxReq),
	}
}

// Acquire will wait untill underlying channel has free space.
func (s *Semaphore) Acquire() {
	s.semaCh <- struct{}{}
}

// Release must be used after Require() to free the queue.
func (s *Semaphore) Release() {
	<-s.semaCh
}
