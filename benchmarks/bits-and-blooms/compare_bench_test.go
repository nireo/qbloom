package bitsandbloomsbench

import (
	"testing"

	bitsbloom "github.com/bits-and-blooms/bloom/v3"
	qbloom "github.com/nireo/qbloom"
)

var benchmarkBoolSink bool

// bits-and-blooms exposes TestAndAdd* for the same "check then insert" behavior as Filter.Add*.
func BenchmarkCompareStringTestAndAdd(b *testing.B) {
	values := benchmarkStrings()
	reference := qbloom.NewForSeeded(len(values), 0.01, 1)
	numBits := uint(reference.NumBits())
	numHashes := uint(reference.NumHashes())

	b.Run("qbloom", func(b *testing.B) {
		f := qbloom.NewSeeded(int(numBits), int(numHashes), 1)
		b.ReportAllocs()
		b.ResetTimer()

		var result bool
		for i := 0; i < b.N; i++ {
			result = f.AddString(values[i%len(values)])
		}
		benchmarkBoolSink = result
	})

	b.Run("bits-and-blooms", func(b *testing.B) {
		f := bitsbloom.New(numBits, numHashes)
		b.ReportAllocs()
		b.ResetTimer()

		var result bool
		for i := 0; i < b.N; i++ {
			result = f.TestAndAddString(values[i%len(values)])
		}
		benchmarkBoolSink = result
	})
}

func BenchmarkCompareStringContains(b *testing.B) {
	values := benchmarkStrings()
	reference := qbloom.NewForSeeded(len(values), 0.01, 1)
	numBits := uint(reference.NumBits())
	numHashes := uint(reference.NumHashes())

	b.Run("qbloom", func(b *testing.B) {
		f := qbloom.NewSeeded(int(numBits), int(numHashes), 1)
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
	})

	b.Run("bits-and-blooms", func(b *testing.B) {
		f := bitsbloom.New(numBits, numHashes)
		for _, value := range values {
			f.AddString(value)
		}
		b.ReportAllocs()
		b.ResetTimer()

		var result bool
		for i := 0; i < b.N; i++ {
			result = f.TestString(values[i%len(values)])
		}
		benchmarkBoolSink = result
	})
}

func BenchmarkCompareBytesTestAndAdd(b *testing.B) {
	values := benchmarkBytes()
	reference := qbloom.NewForSeeded(len(values), 0.01, 1)
	numBits := uint(reference.NumBits())
	numHashes := uint(reference.NumHashes())

	b.Run("qbloom", func(b *testing.B) {
		f := qbloom.NewSeeded(int(numBits), int(numHashes), 1)
		b.ReportAllocs()
		b.ResetTimer()

		var result bool
		for i := 0; i < b.N; i++ {
			result = f.Add(values[i%len(values)])
		}
		benchmarkBoolSink = result
	})

	b.Run("bits-and-blooms", func(b *testing.B) {
		f := bitsbloom.New(numBits, numHashes)
		b.ReportAllocs()
		b.ResetTimer()

		var result bool
		for i := 0; i < b.N; i++ {
			result = f.TestAndAdd(values[i%len(values)])
		}
		benchmarkBoolSink = result
	})
}

func BenchmarkCompareBytesContains(b *testing.B) {
	values := benchmarkBytes()
	reference := qbloom.NewForSeeded(len(values), 0.01, 1)
	numBits := uint(reference.NumBits())
	numHashes := uint(reference.NumHashes())

	b.Run("qbloom", func(b *testing.B) {
		f := qbloom.NewSeeded(int(numBits), int(numHashes), 1)
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
	})

	b.Run("bits-and-blooms", func(b *testing.B) {
		f := bitsbloom.New(numBits, numHashes)
		for _, value := range values {
			f.Add(value)
		}
		b.ReportAllocs()
		b.ResetTimer()

		var result bool
		for i := 0; i < b.N; i++ {
			result = f.Test(values[i%len(values)])
		}
		benchmarkBoolSink = result
	})
}
