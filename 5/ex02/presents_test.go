package main

import "testing"

func TestGetNCoolestPresents(t *testing.T) {
	presents := ph{{5, 1}, {4, 5}, {3, 1}, {5, 2}}

	got, _ := getNCoolestPresents(presents, 2)
	expect := []Present{{5, 1}, {5, 2}}

	for i, val := range got {
		if expect[i] != val {
			t.Error("Wrong answer")
		}
	}
}
