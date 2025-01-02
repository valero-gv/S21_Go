package main

type TreeNode struct {
	HasToy bool
	Left   *TreeNode
	Right  *TreeNode
}

func countToys(node *TreeNode) int {
	if node == nil {
		return 0
	}
	count := 0
	if node.HasToy {
		count = 1
	}
	count += countToys(node.Left)
	count += countToys(node.Right)
	return count
}

func areToysBalanced(root *TreeNode) bool {
	if root == nil {
		return true
	}
	countLeft := countToys(root.Left)
	countRight := countToys(root.Right)
	return countRight == countLeft
}
