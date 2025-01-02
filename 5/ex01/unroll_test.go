package main

import (
	"testing"
)

func TestUnroll(t *testing.T) {
	tree := TreeNode{
		HasToy: true,
		Left: &TreeNode{
			HasToy: true,
			Left: &TreeNode{
				HasToy: true,
				Left:   nil,
				Right:  nil,
			},
			Right: &TreeNode{
				HasToy: false,
				Left:   nil,
				Right:  nil,
			},
		},
		Right: &TreeNode{
			HasToy: false,
			Left: &TreeNode{
				HasToy: true,
				Left:   nil,
				Right:  nil,
			},
			Right: &TreeNode{
				HasToy: true,
				Left:   nil,
				Right:  nil,
			},
		},
	}

	got := unrollGarland(&tree)
	expect := []bool{true, true, false, true, true, false, true}
	for i, val := range got {
		if expect[i] != val {
			t.Error("Got not expected slice")
		}
	}
}
