package main

import (
	"testing"

	"github.com/matryer/is"
)

func Test_SpeedTestFrequency(t *testing.T) {
	is := is.New(t)
	minutesInDay := 1440
	maxTests := 6

	//
	tf := func() int {
		numTests := 0
		for i := 1; i < minutesInDay; {
			t1 := speedTestInterval(maxTests)
			i += int(t1.Minutes())
			numTests++
		}

		return numTests
	}

	results := []int{tf(), tf(), tf(), tf(), tf(), tf(), tf(), tf(), tf()}
	is.True(max(results) <= maxTests+1)
	is.True(min(results) > maxTests/2)
}

func Test_MaxSpeedTestFrequency(t *testing.T) {
	is := is.New(t)
	minutesInDay := 1440
	maxTests := 32

	//
	tf := func() int {
		numTests := 0
		for i := 1; i < minutesInDay; {
			t1 := speedTestInterval(maxTests)
			i += int(t1.Minutes())
			numTests++
		}

		return numTests
	}

	results := []int{tf(), tf(), tf(), tf(), tf(), tf(), tf(), tf(), tf()}
	is.True(max(results) <= maxTests+2)
	is.True(min(results) > maxTests/2)
}

func Test_MinSpeedTestFrequency(t *testing.T) {
	is := is.New(t)
	minutesInDay := 1440
	minTests := 1

	//
	tf := func() int {
		numTests := 0
		for i := 1; i < minutesInDay; {
			t1 := speedTestInterval(minTests)
			i += int(t1.Minutes())
			numTests++
		}

		return numTests
	}

	results := []int{tf(), tf(), tf(), tf(), tf(), tf(), tf(), tf(), tf()}
	is.True(max(results) <= minTests+1)
}
