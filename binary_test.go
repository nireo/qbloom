package qbloom

import (
	"encoding/binary"
	"strconv"
	"testing"
)

func TestFilterBinaryRoundTrip(t *testing.T) {
	original := NewForSeeded(2000, 0.01, 42)
	for i := 0; i < 2000; i++ {
		original.AddString("member-" + strconv.Itoa(i))
	}

	data, err := original.MarshalBinary()
	if err != nil {
		t.Fatalf("MarshalBinary() error = %v", err)
	}

	var decoded Filter
	if err := decoded.UnmarshalBinary(data); err != nil {
		t.Fatalf("UnmarshalBinary() error = %v", err)
	}

	if decoded.numBits != original.numBits {
		t.Fatalf("numBits mismatch: got %d want %d", decoded.numBits, original.numBits)
	}
	if decoded.numHashes != original.numHashes {
		t.Fatalf("numHashes mismatch: got %d want %d", decoded.numHashes, original.numHashes)
	}
	if decoded.seed != original.seed {
		t.Fatalf("seed mismatch: got %d want %d", decoded.seed, original.seed)
	}
	if !equalWords(decoded.words, original.words) {
		t.Fatalf("word data mismatch after round trip")
	}

	for i := 0; i < 2000; i++ {
		value := "member-" + strconv.Itoa(i)
		if !decoded.ContainsString(value) {
			t.Fatalf("missing inserted value %q after round trip", value)
		}
	}
}

func TestFilterUnmarshalBinaryRejectsInvalidData(t *testing.T) {
	original := NewSeeded(1024, 4, 7)
	data, err := original.MarshalBinary()
	if err != nil {
		t.Fatalf("MarshalBinary() error = %v", err)
	}

	corrupt := append([]byte(nil), data...)
	binary.LittleEndian.PutUint64(corrupt[8:16], 65)

	var decoded Filter
	if err := decoded.UnmarshalBinary(corrupt); err == nil {
		t.Fatalf("UnmarshalBinary() unexpectedly succeeded for corrupt data")
	}
}

func TestAtomicFilterBinaryRoundTrip(t *testing.T) {
	original := NewAtomicForSeeded(2000, 0.01, 42)
	for i := 0; i < 2000; i++ {
		original.AddString("member-" + strconv.Itoa(i))
	}

	data, err := original.MarshalBinary()
	if err != nil {
		t.Fatalf("MarshalBinary() error = %v", err)
	}

	var decoded AtomicFilter
	if err := decoded.UnmarshalBinary(data); err != nil {
		t.Fatalf("UnmarshalBinary() error = %v", err)
	}

	if decoded.numBits != original.numBits {
		t.Fatalf("numBits mismatch: got %d want %d", decoded.numBits, original.numBits)
	}
	if decoded.numHashes != original.numHashes {
		t.Fatalf("numHashes mismatch: got %d want %d", decoded.numHashes, original.numHashes)
	}
	if decoded.seed != original.seed {
		t.Fatalf("seed mismatch: got %d want %d", decoded.seed, original.seed)
	}
	if !equalWords(snapshotAtomicWords(&decoded), snapshotAtomicWords(original)) {
		t.Fatalf("word data mismatch after round trip")
	}

	for i := 0; i < 2000; i++ {
		value := "member-" + strconv.Itoa(i)
		if !decoded.ContainsString(value) {
			t.Fatalf("missing inserted value %q after round trip", value)
		}
	}
}

func TestAtomicFilterUnmarshalBinaryRejectsInvalidData(t *testing.T) {
	original := NewAtomicSeeded(1024, 4, 7)
	data, err := original.MarshalBinary()
	if err != nil {
		t.Fatalf("MarshalBinary() error = %v", err)
	}

	corrupt := append([]byte(nil), data...)
	binary.LittleEndian.PutUint64(corrupt[8:16], 65)

	var decoded AtomicFilter
	if err := decoded.UnmarshalBinary(corrupt); err == nil {
		t.Fatalf("UnmarshalBinary() unexpectedly succeeded for corrupt data")
	}
}
