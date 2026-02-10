package layout

import "testing"

func TestSum(t *testing.T) {
	t.Parallel()
	testSum(t, []int{}, 0)
	testSum(t, []int{0}, 0)
	testSum(t, []int{1}, 1)
	testSum(t, []int{12, 34}, 46)
	testSum(t, []int{12, 34, 56}, 102)
}

func testSum[N number](t *testing.T, s []N, want N) {
	n := sum(s)
	if n != want {
		t.Errorf("sum(%v) = %v; want %v", s, n, want)
	}
}
