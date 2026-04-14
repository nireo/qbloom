package qbloom

import (
	"sync"
	"sync/atomic"
	"testing"

	bitsbloom "github.com/bits-and-blooms/bloom/v3"
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

func BenchmarkAtomicAddString(b *testing.B) {
	values := benchmarkStrings()
	f := NewAtomicForSeeded(len(values), 0.01, 1)
	b.ReportAllocs()
	b.ResetTimer()

	var result bool
	for i := 0; i < b.N; i++ {
		result = f.AddString(values[i%len(values)])
	}
	benchmarkBoolSink = result
}

func BenchmarkAtomicContainsString(b *testing.B) {
	values := benchmarkStrings()
	f := NewAtomicForSeeded(len(values), 0.01, 1)
	for _, value := range values {
		f.AddString(value)
	}
	b.ReportAllocs()
	b.ResetTimer()

	var result bool
	for i := 0; i < b.N; i++ {
		result = f.ContainsString(values[i%len(values)])
	}
	benchmarkBoolSink = result
}

func BenchmarkAtomicAddBytes(b *testing.B) {
	values := benchmarkBytes()
	f := NewAtomicForSeeded(len(values), 0.01, 1)
	b.ReportAllocs()
	b.ResetTimer()

	var result bool
	for i := 0; i < b.N; i++ {
		result = f.Add(values[i%len(values)])
	}
	benchmarkBoolSink = result
}

func BenchmarkAtomicContainsBytes(b *testing.B) {
	values := benchmarkBytes()
	f := NewAtomicForSeeded(len(values), 0.01, 1)
	for _, value := range values {
		f.Add(value)
	}
	b.ReportAllocs()
	b.ResetTimer()

	var result bool
	for i := 0; i < b.N; i++ {
		result = f.Contains(values[i%len(values)])
	}
	benchmarkBoolSink = result
}

func BenchmarkAtomicAddHash(b *testing.B) {
	hashes := benchmarkHashes(1)
	f := NewAtomicForSeeded(len(hashes), 0.01, 1)
	b.ReportAllocs()
	b.ResetTimer()

	var result bool
	for i := 0; i < b.N; i++ {
		result = f.AddHash(hashes[i%len(hashes)])
	}
	benchmarkBoolSink = result
}

func BenchmarkAtomicContainsHash(b *testing.B) {
	hashes := benchmarkHashes(1)
	f := NewAtomicForSeeded(len(hashes), 0.01, 1)
	for _, hash := range hashes {
		f.AddHash(hash)
	}
	b.ReportAllocs()
	b.ResetTimer()

	var result bool
	for i := 0; i < b.N; i++ {
		result = f.ContainsHash(hashes[i%len(hashes)])
	}
	benchmarkBoolSink = result
}

func BenchmarkCompareAtomicParallelStringAdd(b *testing.B) {
	values := benchmarkStrings()
	reference := NewForSeeded(len(values), 0.01, 1)
	numBits := uint(reference.NumBits())
	numHashes := uint(reference.NumHashes())

	b.Run("atomic", func(b *testing.B) {
		f := NewAtomicSeeded(int(numBits), int(numHashes), 1)
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
	reference := NewForSeeded(len(values), 0.01, 1)
	numBits := uint(reference.NumBits())
	numHashes := uint(reference.NumHashes())

	b.Run("atomic", func(b *testing.B) {
		f := NewAtomicSeeded(int(numBits), int(numHashes), 1)
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
	reference := NewForSeeded(len(values), 0.01, 1)
	numBits := uint(reference.NumBits())
	numHashes := uint(reference.NumHashes())

	b.Run("atomic", func(b *testing.B) {
		f := NewAtomicSeeded(int(numBits), int(numHashes), 1)
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
	reference := NewForSeeded(len(values), 0.01, 1)
	numBits := uint(reference.NumBits())
	numHashes := uint(reference.NumHashes())

	b.Run("atomic", func(b *testing.B) {
		f := NewAtomicSeeded(int(numBits), int(numHashes), 1)
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

func BenchmarkAtomicParallelAddString(b *testing.B) {
	values := benchmarkStrings()
	f := NewAtomicForSeeded(len(values), 0.01, 1)
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
}

func BenchmarkAtomicParallelContainsString(b *testing.B) {
	values := benchmarkStrings()
	f := NewAtomicForSeeded(len(values), 0.01, 1)
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
}

func BenchmarkAtomicParallelAddHash(b *testing.B) {
	hashes := benchmarkHashes(1)
	f := NewAtomicForSeeded(len(hashes), 0.01, 1)
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		var local uint64
		for pb.Next() {
			if f.AddHash(hashes[i%len(hashes)]) {
				local++
			}
			i++
		}
		benchmarkCountSink.Add(local)
	})
}

func BenchmarkAtomicParallelContainsHash(b *testing.B) {
	hashes := benchmarkHashes(1)
	f := NewAtomicForSeeded(len(hashes), 0.01, 1)
	for _, hash := range hashes {
		f.AddHash(hash)
	}
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		var local uint64
		for pb.Next() {
			if f.ContainsHash(hashes[i%len(hashes)]) {
				local++
			}
			i++
		}
		benchmarkCountSink.Add(local)
	})
}
