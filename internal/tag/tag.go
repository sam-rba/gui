package tag

type Tagged[V, T any] struct {
	Val V
	Tag T
}

func Tag[V, T any](out chan<- Tagged[V, T], in <-chan V, f func(V) T) {
	for val := range in {
		out <- Tagged[V, T]{val, f(val)}
	}
}
