package layout

// WaitGroup is a counting semaphore used to wait for a group of goroutines to finish.
// It differs from sync/WaitGroup in that Wait() is a channel rather than a blocking function.
type waitgroup struct {
	done    chan<- struct{}
	alldone <-chan struct{}
}

// NewWaitGroup creates a group of n goroutines: a semaphore with a capacity of n.
func newWaitGroup(n uint) waitgroup {
	done, alldone := make(chan struct{}), make(chan struct{})
	go func() {
		for ; n > 0; n-- {
			<-done
		}
		alldone <- *new(struct{})
		close(done)
		close(alldone)
	}()
	return waitgroup{done, alldone}
}

// Done decrements the task counter by one.
func (wg waitgroup) Done() {
	wg.done <- *new(struct{})
}

// Wait returns a channel that blocks until the task counter is zero.
func (wg waitgroup) Wait() <-chan struct{} {
	return wg.alldone
}
