package tag

import "context"

type Tagged[V, T any] struct {
	Val V
	Tag T
}

func Tag[V, T any](ctx context.Context, out chan<- Tagged[V, T], in <-chan V, f func(V) T) {
	for {
		select {
		case val, ok := <-in:
			if !ok {
				return
			}
			out <- Tagged[V, T]{val, f(val)}
		case <-ctx.Done():
			return
		}
	}
}
