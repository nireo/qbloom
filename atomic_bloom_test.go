package qbloom

import (
	"fmt"
	"sync"
	"testing"
)

func TestAtomicAddContains(t *testing.T) {
	f := NewAtomic(1024, 4)

	for i := 0; i < 5000; i++ {
		value := fmt.Sprintf("member-%d", i)
		if f.AddString(value) && i == 0 {
			t.Fatalf("first insert reported already present")
		}
	}

	for i := 0; i < 5000; i++ {
		value := fmt.Sprintf("member-%d", i)
		if !f.ContainsString(value) {
			t.Fatalf("missing inserted value %q", value)
		}
	}
}

func TestAtomicRepeatInsert(t *testing.T) {
	f := NewAtomic(1024, 4)

	if f.AddString("hello") {
		t.Fatalf("first insert should report false")
	}
	if !f.AddString("hello") {
		t.Fatalf("second insert should report true")
	}
}

func TestAtomicReset(t *testing.T) {
	f := NewAtomic(4096, 4)

	for i := 0; i < 1000; i++ {
		f.AddString(fmt.Sprintf("member-%d", i))
	}
	f.Reset()

	for i := 0; i < 1000; i++ {
		value := fmt.Sprintf("member-%d", i)
		if f.ContainsString(value) {
			t.Fatalf("value %q still present after reset", value)
		}
	}
}

func TestAtomicSeededFiltersAreDeterministic(t *testing.T) {
	left := NewAtomicForSeeded(2000, 0.01, 42)
	right := NewAtomicForSeeded(2000, 0.01, 42)
	otherSeed := NewAtomicForSeeded(2000, 0.01, 43)

	for i := 0; i < 2000; i++ {
		value := fmt.Sprintf("member-%d", i)
		left.AddString(value)
		right.AddString(value)
		otherSeed.AddString(value)
	}

	if !equalWords(snapshotAtomicWords(left), snapshotAtomicWords(right)) {
		t.Fatalf("same seed produced different filters")
	}
	if equalWords(snapshotAtomicWords(left), snapshotAtomicWords(otherSeed)) {
		t.Fatalf("different seeds produced the same filter")
	}
}

func TestAtomicConcurrentAddContains(t *testing.T) {
	const (
		workers = 8
		total   = 1 << 15
	)

	values := make([]string, total)
	for i := range values {
		values[i] = fmt.Sprintf("member-%d", i)
	}

	f := NewAtomicForSeeded(total, 0.01, 7)
	start := make(chan struct{})
	errCh := make(chan error, workers)
	var wg sync.WaitGroup

	for worker := 0; worker < workers; worker++ {
		worker := worker
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for i := worker; i < total; i += workers {
				value := values[i]
				f.AddString(value)
				if !f.ContainsString(value) {
					errCh <- fmt.Errorf("value %q missing immediately after insert", value)
					return
				}
			}
		}()
	}

	close(start)
	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			t.Fatal(err)
		}
	}

	for _, value := range values {
		if !f.ContainsString(value) {
			t.Fatalf("missing inserted value %q", value)
		}
	}
}

func snapshotAtomicWords(f *AtomicFilter) []uint64 {
	words := make([]uint64, len(f.words))
	for i := range f.words {
		words[i] = f.words[i].Load()
	}
	return words
}
