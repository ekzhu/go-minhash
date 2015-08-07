package minhash

import (
	"math"
	"math/rand"
)

const (
	// Mersenne prime for universal hash functions with 32-bit keys
	// 2^61 - 1
	p32     = uint64(61)
	prime32 = uint64((1 << p32) - 1)
)

type Hash32 func([]byte) uint32

// MinWise is a collection of minimum hashes for a set
type MinWise struct {
	minimums []uint32
	h        Hash32
	a        []uint64
	b        []uint64
}

// NewMinWise returns a new MinWise Hashsing implementation
func NewMinWise(h Hash32, size int, seed int64) *MinWise {

	m := &MinWise{
		minimums: make([]uint32, size),
		h:        h,
		a:        make([]uint64, size),
		b:        make([]uint64, size),
	}
	p := int64(prime32)
	r := rand.New(rand.NewSource(seed))
	for i := 0; i < size; i++ {
		m.minimums[i] = math.MaxUint32
		for {
			a := r.Int63n(p)
			if a != 0 {
				m.a[i] = uint64(a)
				break
			}
		}
		m.b[i] = uint64(r.Int63n(p))
	}
	return m
}

// Push adds an element to the set.
func (m *MinWise) Push(b []byte) {

	var hv, phv uint64
	hv = uint64(m.h(b))
	for i := range m.minimums {
		// Because a, b, and hv are all 32-bit padded to 64-bit
		// we can do multiplication without worrying about overflow.
		phv = m.a[i]*hv + m.b[i]
		// The fast way to compute phv % prime32
		for (phv >> p32) != 0 {
			phv = (phv & prime32) + (phv >> p32)
		}
		// The fast way to compute phv % 2^32
		phv = phv & uint64(math.MaxUint32)
		if uint32(phv) < m.minimums[i] {
			m.minimums[i] = uint32(phv)
		}
	}
}

// Merge combines the signatures of the second set, creating the signature of their union.
func (m *MinWise) Merge(m2 *MinWise) {

	for i, v := range m2.minimums {

		if v < m.minimums[i] {
			m.minimums[i] = v
		}
	}
}

// Cardinality estimates the cardinality of the set
func (m *MinWise) Cardinality() int {

	// http://www.cohenwang.com/edith/Papers/tcest.pdf

	sum := 0.0

	for _, v := range m.minimums {
		sum += -math.Log(float64(math.MaxUint32-v) / float64(math.MaxUint32))
	}

	return int(float64(len(m.minimums)-1) / sum)
}

// Signature returns a signature for the set.
func (m *MinWise) Signature() []uint32 {
	return m.minimums
}

// Similarity computes an estimate for the similarity between the two sets.
func (m *MinWise) Similarity(m2 *MinWise) float64 {

	if len(m.minimums) != len(m2.minimums) {
		panic("minhash minimums size mismatch")
	}

	intersect := 0

	for i := range m.minimums {
		if m.minimums[i] == m2.minimums[i] {
			intersect++
		}
	}

	return float64(intersect) / float64(len(m.minimums))
}

// SignatureBbit returns a b-bit reduction of the signature.  This will result in unused bits at the high-end of the words if b does not divide 32 evenly.
func (m *MinWise) SignatureBbit(b uint) []uint32 {

	var sig []uint32 // full signature
	var w uint32     // current word
	bits := uint(32) // bits free in current word

	mask := uint32(1<<b) - 1

	for _, v := range m.minimums {
		if bits >= b {
			w <<= b
			w |= v & mask
			bits -= b
		} else {
			sig = append(sig, w)
			w = 0
			bits = 32
		}
	}

	if bits != 32 {
		sig = append(sig, w)
	}

	return sig
}

// SimilarityBbit computes an estimate for the similarity between two b-bit signatures
func SimilarityBbit(sig1, sig2 []uint32, b uint) float64 {

	if len(sig1) != len(sig2) {
		panic("signature size mismatch")
	}

	intersect := 0
	count := 0

	mask := uint32(1<<b) - 1

	for i := range sig1 {
		w1 := sig1[i]
		w2 := sig2[i]

		bits := uint(32)

		for bits >= b {
			v1 := (w1 & mask)
			v2 := (w2 & mask)

			count++
			if v1 == v2 {
				intersect++
			}

			bits -= b
			w1 >>= b
			w2 >>= b
		}
	}

	return float64(intersect) / float64(count)
}
