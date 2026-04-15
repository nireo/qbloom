package bitsandbloomsbench

import (
	"strconv"
	"sync"
	"sync/atomic"
	"testing"

	bitsbloom "github.com/bits-and-blooms/bloom/v3"
	qbloom "github.com/nireo/qbloom"
)

var benchmarkCountSink atomic.Uint64

type mutexBloom struct {
	mu sync.Mutex
	f  *bitsbloom.BloomFilter
}

func newMutexBloom(numBits, numHashes uint) *mutexBloom {
	return &mutexBloom{f: bitsbloom.New(numBits, numHashes)}
}

func (b *mutexBloom) addString(s string) {
	b.mu.Lock()
	b.f.AddString(s)
	b.mu.Unlock()
}

func (b *mutexBloom) testAndAddString(s string) bool {
	b.mu.Lock()
	result := b.f.TestAndAddString(s)
	b.mu.Unlock()
	return result
}

func (b *mutexBloom) testString(s string) bool {
	b.mu.Lock()
	result := b.f.TestString(s)
	b.mu.Unlock()
	return result
}

func (b *mutexBloom) addBytes(value []byte) {
	b.mu.Lock()
	b.f.Add(value)
	b.mu.Unlock()
}

func (b *mutexBloom) testAndAddBytes(value []byte) bool {
	b.mu.Lock()
	result := b.f.TestAndAdd(value)
	b.mu.Unlock()
	return result
}

func (b *mutexBloom) testBytes(value []byte) bool {
	b.mu.Lock()
	result := b.f.Test(value)
	b.mu.Unlock()
	return result
}

func BenchmarkCompareAtomicParallelStringAdd(b *testing.B) {
	values := benchmarkStrings()
	reference := qbloom.NewForSeeded(len(values), 0.01, 1)
	numBits := uint(reference.NumBits())
	numHashes := uint(reference.NumHashes())

	b.Run("atomic", func(b *testing.B) {
		f := qbloom.NewAtomicSeeded(int(numBits), int(numHashes), 1)
		b.ReportAllocs()
		b.ResetTimer()

		b.RunParallel(func(pb *testing.PB) {
			i := 0
			var local uint64
			for pb.Next() {
				if f.AddString(values[i%len(values)]) {
					local++
				}
				i++
			}
			benchmarkCountSink.Add(local)
		})
	})

	b.Run("bits-and-blooms-mutex", func(b *testing.B) {
		f := newMutexBloom(numBits, numHashes)
		b.ReportAllocs()
		b.ResetTimer()

		b.RunParallel(func(pb *testing.PB) {
			i := 0
			var local uint64
			for pb.Next() {
				if f.testAndAddString(values[i%len(values)]) {
					local++
				}
				i++
			}
			benchmarkCountSink.Add(local)
		})
	})
}

func BenchmarkCompareAtomicParallelStringContains(b *testing.B) {
	values := benchmarkStrings()
	reference := qbloom.NewForSeeded(len(values), 0.01, 1)
	numBits := uint(reference.NumBits())
	numHashes := uint(reference.NumHashes())

	b.Run("atomic", func(b *testing.B) {
		f := qbloom.NewAtomicSeeded(int(numBits), int(numHashes), 1)
		for _, value := range values {
			f.AddString(value)
		}
		b.ReportAllocs()
		b.ResetTimer()

		b.RunParallel(func(pb *testing.PB) {
			i := 0
			var local uint64
			for pb.Next() {
				if f.ContainsString(values[i%len(values)]) {
					local++
				}
				i++
			}
			benchmarkCountSink.Add(local)
		})
	})

	b.Run("bits-and-blooms-mutex", func(b *testing.B) {
		f := newMutexBloom(numBits, numHashes)
		for _, value := range values {
			f.addString(value)
		}
		b.ReportAllocs()
		b.ResetTimer()

		b.RunParallel(func(pb *testing.PB) {
			i := 0
			var local uint64
			for pb.Next() {
				if f.testString(values[i%len(values)]) {
					local++
				}
				i++
			}
			benchmarkCountSink.Add(local)
		})
	})
}

func BenchmarkCompareAtomicParallelBytesAdd(b *testing.B) {
	values := benchmarkBytes()
	reference := qbloom.NewForSeeded(len(values), 0.01, 1)
	numBits := uint(reference.NumBits())
	numHashes := uint(reference.NumHashes())

	b.Run("atomic", func(b *testing.B) {
		f := qbloom.NewAtomicSeeded(int(numBits), int(numHashes), 1)
		b.ReportAllocs()
		b.ResetTimer()

		b.RunParallel(func(pb *testing.PB) {
			i := 0
			var local uint64
			for pb.Next() {
				if f.Add(values[i%len(values)]) {
					local++
				}
				i++
			}
			benchmarkCountSink.Add(local)
		})
	})

	b.Run("bits-and-blooms-mutex", func(b *testing.B) {
		f := newMutexBloom(numBits, numHashes)
		b.ReportAllocs()
		b.ResetTimer()

		b.RunParallel(func(pb *testing.PB) {
			i := 0
			var local uint64
			for pb.Next() {
				if f.testAndAddBytes(values[i%len(values)]) {
					local++
				}
				i++
			}
			benchmarkCountSink.Add(local)
		})
	})
}

func BenchmarkCompareAtomicParallelBytesContains(b *testing.B) {
	values := benchmarkBytes()
	reference := qbloom.NewForSeeded(len(values), 0.01, 1)
	numBits := uint(reference.NumBits())
	numHashes := uint(reference.NumHashes())

	b.Run("atomic", func(b *testing.B) {
		f := qbloom.NewAtomicSeeded(int(numBits), int(numHashes), 1)
		for _, value := range values {
			f.Add(value)
		}
		b.ReportAllocs()
		b.ResetTimer()

		b.RunParallel(func(pb *testing.PB) {
			i := 0
			var local uint64
			for pb.Next() {
				if f.Contains(values[i%len(values)]) {
					local++
				}
				i++
			}
			benchmarkCountSink.Add(local)
		})
	})

	b.Run("bits-and-blooms-mutex", func(b *testing.B) {
		f := newMutexBloom(numBits, numHashes)
		for _, value := range values {
			f.addBytes(value)
		}
		b.ReportAllocs()
		b.ResetTimer()

		b.RunParallel(func(pb *testing.PB) {
			i := 0
			var local uint64
			for pb.Next() {
				if f.testBytes(values[i%len(values)]) {
					local++
				}
				i++
			}
			benchmarkCountSink.Add(local)
		})
	})
}

func benchmarkStrings() []string {
	const size = 1 << 16
	values := make([]string, size)
	for i := range values {
		values[i] = "member-" + strconv.Itoa(i)
	}
	return values
}

func benchmarkBytes() [][]byte {
	strings := benchmarkStrings()
	values := make([][]byte, len(strings))
	for i, value := range strings {
		values[i] = []byte(value)
	}
	return values
}
