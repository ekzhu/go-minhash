package minhash

import (
	"fmt"
	"hash/fnv"
	"math"
	"testing"
)

var hash32 Hash32

func init() {
	fnvHash32 := fnv.New32a()
	hash32 = func(b []byte) uint32 {
		fnvHash32.Reset()
		fnvHash32.Write(b)
		return fnvHash32.Sum32()
	}
}

func data(size int) [][]byte {
	d := make([][]byte, size)
	for i := range d {
		d[i] = []byte(fmt.Sprintf("salt%d %d", i, size))
	}
	return d
}

func hashing(mh *MinWise, start, end int, data [][]byte) {
	for i := start; i < end; i++ {
		mh.Push(data[i])
	}
}

func benchmark(minhashSize, dataSize int) {
	if dataSize < 10 {
		fmt.Printf("\n")
		return
	}
	// Data is a set of unique values
	d := data(dataSize)
	// a and b are two subsets of data with some overlaps
	a_start, a_end := 0, int(float64(dataSize)*0.65)
	b_start, b_end := int(float64(dataSize)*0.35), dataSize

	m1 := NewMinWise(hash32, minhashSize, 0)
	m2 := NewMinWise(hash32, minhashSize, 0)

	hashing(m1, a_start, a_end, d)
	hashing(m2, b_start, b_end, d)

	est := m1.Similarity(m2)
	act := float64(a_end-b_start) / float64(b_end-a_start)
	err := math.Abs(act - est)
	fmt.Printf("Data size: %8d, ", dataSize)
	fmt.Printf("Real: %.8f, ", act)
	fmt.Printf("Estimated: %.8f, ", est)
	fmt.Printf("Error: %.8f\n", err)
}

func BenchmarkMinWise64(b *testing.B) {
	benchmark(64, b.N)
}

func BenchmarkMinWise128(b *testing.B) {
	benchmark(128, b.N)
}

func BenchmarkMinWise256(b *testing.B) {
	benchmark(256, b.N)
}

func BenchmarkMinWise512(b *testing.B) {
	benchmark(512, b.N)
}
