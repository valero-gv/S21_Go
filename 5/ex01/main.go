package main

type TreeNode struct {
	HasToy bool
	Left   *TreeNode
	Right  *TreeNode
}

func unrollGarland(root *TreeNode) []bool {
	if root == nil {
		return nil
	}

	var result []bool
	queue := []*TreeNode{root}
	zigzag := false

	for len(queue) > 0 {
		levelSize := len(queue)
		levelNodes := make([]bool, levelSize)

		for i := 0; i < levelSize; i++ {
			node := queue[i]

			if zigzag {
				levelNodes[levelSize-1-i] = node.HasToy
			} else {
				levelNodes[i] = node.HasToy
			}

			if node.Right != nil {
				queue = append(queue, node.Right)
			}
			if node.Left != nil {
				queue = append(queue, node.Left)
			}
		}

		result = append(result, levelNodes...)
		queue = queue[levelSize:]
		zigzag = !zigzag
	}

	return result
}
