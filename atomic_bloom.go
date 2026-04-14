package qbloom

import (
	"math"
	"math/bits"
	"sync/atomic"

	"github.com/zeebo/xxh3"
)

// AtomicFilter is a concurrency-safe Bloom filter backed by atomic 64-bit words.
// False positives are possible, but false negatives are not.
type AtomicFilter struct {
	words     []atomic.Uint64
	numBits   uint64
	numHashes uint32
	seed      uint64
}

// NewAtomic returns a concurrency-safe filter with the requested bit count and hash count.
// The bit count is rounded up to a whole number of 64-bit words.
func NewAtomic(numBits, numHashes int) *AtomicFilter {
	return NewAtomicSeeded(numBits, numHashes, 0)
}

// NewAtomicSeeded returns a concurrency-safe filter with the requested bit count,
// hash count, and xxh3 seed.
func NewAtomicSeeded(numBits, numHashes int, seed uint64) *AtomicFilter {
	if numBits <= 0 {
		panic("fastbloom: numBits must be > 0")
	}

	words := wordCount(numBits)
	return &AtomicFilter{
		words:     make([]atomic.Uint64, words),
		numBits:   uint64(words * wordBits),
		numHashes: normalizeHashes(numHashes),
		seed:      seed,
	}
}

// NewAtomicFor returns a concurrency-safe filter sized for expectedItems and the
// target false positive rate.
func NewAtomicFor(expectedItems int, falsePositiveRate float64) *AtomicFilter {
	return NewAtomicForSeeded(expectedItems, falsePositiveRate, 0)
}

// NewAtomicForSeeded returns a concurrency-safe filter sized for expectedItems,
// the target false positive rate, and the provided xxh3 seed.
func NewAtomicForSeeded(expectedItems int, falsePositiveRate float64, seed uint64) *AtomicFilter {
	if math.IsNaN(falsePositiveRate) || falsePositiveRate <= 0 || falsePositiveRate >= 1 {
		panic("fastbloom: falsePositiveRate must be in (0, 1)")
	}

	expectedItems = normalizeItems(expectedItems)
	numBits := optimalBits(expectedItems, falsePositiveRate)
	words := wordCount(numBits)
	actualBits := words * wordBits

	return &AtomicFilter{
		words:     make([]atomic.Uint64, words),
		numBits:   uint64(actualBits),
		numHashes: optimalHashes(actualBits, expectedItems),
		seed:      seed,
	}
}

// NumBits returns the number of addressable bits in the filter.
func (f *AtomicFilter) NumBits() int {
	return int(f.numBits)
}

// NumHashes returns the number of hashes checked per item.
func (f *AtomicFilter) NumHashes() int {
	return int(f.numHashes)
}

// Add inserts b and reports whether it may have already been present.
func (f *AtomicFilter) Add(b []byte) bool {
	return f.AddHash(xxh3.HashSeed(b, f.seed))
}

// Contains reports whether b may be present in the filter.
func (f *AtomicFilter) Contains(b []byte) bool {
	return f.ContainsHash(xxh3.HashSeed(b, f.seed))
}

// AddString inserts s and reports whether it may have already been present.
func (f *AtomicFilter) AddString(s string) bool {
	return f.AddHash(xxh3.HashStringSeed(s, f.seed))
}

// ContainsString reports whether s may be present in the filter.
func (f *AtomicFilter) ContainsString(s string) bool {
	return f.ContainsHash(xxh3.HashStringSeed(s, f.seed))
}

// AddHash inserts a precomputed source hash and reports whether it may have already been present.
func (f *AtomicFilter) AddHash(hash uint64) bool {
	idx := index(f.numBits, hash)
	previouslyContained := f.set(idx)
	if f.numHashes == 1 {
		return previouslyContained
	}

	h1 := hash
	h2 := hash * derivedHashMultiplier
	for i := uint32(1); i < f.numHashes; i++ {
		h1 = bits.RotateLeft64(h1, 5) + h2
		wasSet := f.set(index(f.numBits, h1))
		previouslyContained = previouslyContained && wasSet
	}

	return previouslyContained
}

// ContainsHash reports whether a precomputed source hash may be present in the filter.
func (f *AtomicFilter) ContainsHash(hash uint64) bool {
	if !f.check(index(f.numBits, hash)) {
		return false
	}
	if f.numHashes == 1 {
		return true
	}

	h1 := hash
	h2 := hash * derivedHashMultiplier
	for i := uint32(1); i < f.numHashes; i++ {
		h1 = bits.RotateLeft64(h1, 5) + h2
		if !f.check(index(f.numBits, h1)) {
			return false
		}
	}

	return true
}

// Reset clears all bits in the filter.
func (f *AtomicFilter) Reset() {
	for i := range f.words {
		f.words[i].Store(0)
	}
}

func (f *AtomicFilter) check(idx uint64) bool {
	word := int(idx >> 6)
	mask := uint64(1) << (idx & 63)
	return f.words[word].Load()&mask != 0
}

func (f *AtomicFilter) set(idx uint64) bool {
	word := int(idx >> 6)
	mask := uint64(1) << (idx & 63)
	return f.words[word].Or(mask)&mask != 0
}
