package gui

// sharedVal is a concurrent interface to a piece of shared data.
//
// A client can read the data by sending a channel via request, and the stored value will
// be sent back via the channel. The client is responsible for closing the channel.
//
// The stored value can be changed by sending the new value via set. Requests block until
// the first value is received on set.
//
// A sharedVal should be closed after use.
type sharedVal[T any] struct {
	request chan<- chan T
	set     chan<- T
}

func newSharedVal[T any]() sharedVal[T] {
	request := make(chan chan T)
	set := make(chan T)
	go func() {
		val := <-set // wait for initial value
		for {
			select {
			case v, ok := <-set:
				if !ok { // closed
					return
				}
				val = v
			case req, ok := <-request:
				if !ok { // closed
					return
				}
				go func() { // don't wait for client to receive
					req <- val
				}()
			}
		}
	}()
	return sharedVal[T]{request, set}
}

// get makes a synchronous request and returns the stored value.
func (sv sharedVal[T]) get() T {
	c := make(chan T)
	defer close(c)
	sv.request <- c
	return <-c
}

func (sv sharedVal[T]) close() {
	close(sv.request)
	close(sv.set)
}
