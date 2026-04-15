package qbloom

import (
	"sync/atomic"
	"testing"
)

var benchmarkCountSink atomic.Uint64

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
