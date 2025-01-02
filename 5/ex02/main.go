package main

import (
	"container/heap"
	"fmt"
)

type Present struct {
	Value int
	Size  int
}

type ph []Present

func (p *ph) Push(x interface{}) {
	//TODO implement me
	*p = append(*p, x.(Present))
}

func (p *ph) Pop() any {
	//TODO implement me
	old := *p
	n := len(old)
	x := old[n-1]
	*p = old[:n-1]
	return x
}

func (p ph) Len() int {
	return len(p)
}

func (p ph) Less(i, j int) bool {
	if p[i].Value == p[j].Value {
		return true
	} else if p[i].Value == p[j].Value {
		return p[i].Size < p[j].Size
	}
	return false
}

func (p ph) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func getNCoolestPresents(p []Present, n int) ([]Present, error) {
	if n < 0 || n > len(p) {
		return nil, fmt.Errorf("n must be between 0 and %d, got %d", len(p), n)
	}
	h := &ph{}
	heap.Init(h)

	for _, p := range p {
		heap.Push(h, p)
	}

	coolest := make([]Present, 0, n)
	for i := 0; i < n; i++ {
		if h.Len() == 0 {
			break
		}
		coolest = append(coolest, heap.Pop(h).(Present))
	}

	return coolest, nil
}

func main() {

}
