package layout

type number interface {
	complex | float | integer
}

type complex interface {
	~complex64 | ~complex128
}

type float interface {
	~float32 | ~float64
}

type integer interface {
	signed | unsigned
}

type signed interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

type unsigned interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

func sum[N number](ns []N) N {
	var n N
	for i := range ns {
		n += ns[i]
	}
	return n
}
