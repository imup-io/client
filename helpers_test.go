package main

import (
	"testing"

	"github.com/matryer/is"
	"golang.org/x/exp/constraints"
)

func Test_SpeedTestFrequency(t *testing.T) {
	is := is.New(t)
	hours := 24
	maxTests := 6

	//
	tf := func() int {
		numTests := 0
		for i := 1; i <= hours; {
			t1 := speedTestInterval()
			i += int(t1.Hours())
			numTests++
		}

		return numTests
	}

	results := []int{tf(), tf(), tf(), tf(), tf(), tf(), tf(), tf(), tf()}
	is.True(max(results) <= maxTests)
}

func max[T constraints.Ordered](s []T) T {
	if len(s) == 0 {
		var zero T
		return zero
	}
	m := s[0]
	for _, v := range s {
		if m < v {
			m = v
		}
	}
	return m
}
