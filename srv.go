package gui

// server is a concurrent interface that stores a value and serves it to clients.
// The stored value can be changed by sending the new value via update. server blocks
// until the first value is received.
//
// A client can read the stored value by sending a channel via request. The server
// responds by sending the value back via the channel. The client is responsible for
// closing the channel.
//
// A server should be closed after use.
type server[T any] struct {
	request chan<- chan T
	update  chan<- T
}

func newServer[T any]() server[T] {
	request := make(chan chan T)
	update := make(chan T)
	go func() {
		val := <-update // wait for initial value
		for {
			select {
			case v, ok := <-update:
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
	return server[T]{request, update}
}

// get makes a synchronous request to the server and returns the stored value.
func (srv server[T]) get() T {
	c := make(chan T)
	srv.request <- c
	return <-c
}

func (srv server[T]) close() {
	close(srv.request)
	close(srv.update)
}
