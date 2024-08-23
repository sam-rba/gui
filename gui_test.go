package gui

import "time"

const timeout = 1 * time.Second

// trySend returns true if v can be sent to c within timeout, or false otherwise.
func trySend[T any](c chan<- T, v T, timeout time.Duration) bool {
	timer := time.NewTimer(timeout)
	select {
	case c <- v:
		return true
	case <-timer.C:
		return false
	}
}

// tryRecv returns the value received from c, or false if no value is received within timeout.
func tryRecv[T any](c <-chan T, timeout time.Duration) (*T, bool) {
	timer := time.NewTimer(timeout)
	select {
	case v := <-c:
		return &v, true
	case <-timer.C:
		return nil, false
	}
}
