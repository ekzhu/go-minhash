package minhash

import (
	"container/heap"
	"hash"
	"sort"
)

type intHeap []uint64

func (h intHeap) Len() int { return len(h) }

// actually Greater, since we want a max-heap
func (h intHeap) Less(i, j int) bool { return h[i] > h[j] }
func (h intHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *intHeap) Push(x interface{}) {
	*h = append(*h, x.(uint64))
}

func (h *intHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

type BottomK struct {
	size     int
	h        hash.Hash64
	minimums *intHeap
}

// New returns a new BottomK implementation
func NewBottomK(h hash.Hash64, k int) *BottomK {
	return &BottomK{
		size:     k,
		h:        h,
		minimums: &intHeap{},
	}
}

func (m *BottomK) Add(b []byte) {

	m.h.Reset()
	m.h.Write(b)
	i64 := m.h.Sum64()

	if i64 == 0 {
		return
	}

	if len(*m.minimums) < m.size {
		heap.Push(m.minimums, i64)
		return
	}

	if i64 < (*m.minimums)[0] {
		heap.Pop(m.minimums)
		heap.Push(m.minimums, i64)
	}
}

func (m *BottomK) Signature() []uint64 {
	mins := make(intHeap, len(*m.minimums))
	copy(mins, *m.minimums)
	sort.Sort(mins)
	return mins
}

func (m *BottomK) Similarity(m2 *BottomK) float64 {

	if m.size != m2.size {
		panic("minhash minimums size mismatch")
	}

	mins := make(map[uint64]int, len(*m.minimums))

	for _, v := range *m.minimums {
		mins[v]++
	}

	intersect := 0

	for _, v := range *m2.minimums {
		if count, ok := mins[v]; ok && count > 0 {
			intersect++
			mins[v] = count - 1
		}
	}

	maxlength := len(*m.minimums)
	if maxlength < len(*m2.minimums) {
		maxlength = len(*m2.minimums)
	}

	return float64(intersect) / float64(maxlength)
}