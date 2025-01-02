package main

import (
	"testing"
)

func TestGrabPresents(t *testing.T) {
	presents := ph{{5, 1}, {4, 5}, {3, 1}, {5, 2}}

	got := grabPresents(presents, 3)
	expect := ph{{5, 2}, {5, 1}}

	for i, val := range got {
		if expect[i] != val {
			t.Error("Wrong answer for capacity = 3")
		}
	}

	got = grabPresents(presents, 8)
	expect = ph{{5, 2}, {4, 5}, {5, 1}}

	for i, val := range got {
		if expect[i] != val {
			t.Error("Wrong answer for capacity = 3")
		}
	}
}
