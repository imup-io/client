package main

import (
	"testing"

	"github.com/matryer/is"
)

func Test_SpeedTestFrequency(t *testing.T) {
	is := is.New(t)
	hours := 24
	maxTests := 8

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
	is.True(min(results) > maxTests/2)

}
