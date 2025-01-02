package main

type Present struct {
	Value int
	Size  int
}

type ph []Present

func (p ph) Len() int {
	return len(p)
}

func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func grabPresents(p ph, capacity int) ph {
	pack := make([][]int, len(p)+1)
	for i := range pack {
		pack[i] = make([]int, capacity+1)
	}
	//pack[0][0] = 1
	for i := 1; i <= len(p); i++ {
		for w := 1; w <= capacity; w++ {
			if p[i-1].Size <= w {
				pack[i][w] = Max(pack[i-1][w-p[i-1].Size]+p[i-1].Value, pack[i-1][w])
			} else {
				pack[i][w] = pack[i-1][w]
			}
		}
	}
	k := capacity
	answer := make(ph, 0)
	for i := len(p); i > 0; i-- {
		if pack[i][k] != pack[i-1][k] {
			answer = append(answer, p[i-1])
			k -= p[i-1].Size
		}
	}
	return answer
}
