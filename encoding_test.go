package qbloom

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"encoding/json"
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

	assertFilterState(t, decoded.numBits, decoded.numHashes, decoded.seed, decoded.words, original.numBits, original.numHashes, original.seed, original.words)
	assertContainsMembers(t, decoded.ContainsString)
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

func TestFilterJSONRoundTrip(t *testing.T) {
	original := NewForSeeded(2000, 0.01, 42)
	for i := 0; i < 2000; i++ {
		original.AddString("member-" + strconv.Itoa(i))
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var decoded Filter
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	assertFilterState(t, decoded.numBits, decoded.numHashes, decoded.seed, decoded.words, original.numBits, original.numHashes, original.seed, original.words)
	assertContainsMembers(t, decoded.ContainsString)
}

func TestFilterUnmarshalJSONRejectsInvalidData(t *testing.T) {
	var decoded Filter
	err := json.Unmarshal([]byte(`{"numBits":65,"numHashes":4,"seed":7,"words":[1]}`), &decoded)
	if err == nil {
		t.Fatalf("json.Unmarshal() unexpectedly succeeded for corrupt data")
	}
}

func TestFilterGobRoundTrip(t *testing.T) {
	original := NewForSeeded(2000, 0.01, 42)
	for i := 0; i < 2000; i++ {
		original.AddString("member-" + strconv.Itoa(i))
	}

	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(original); err != nil {
		t.Fatalf("gob Encode() error = %v", err)
	}

	var decoded Filter
	if err := gob.NewDecoder(&buf).Decode(&decoded); err != nil {
		t.Fatalf("gob Decode() error = %v", err)
	}

	assertFilterState(t, decoded.numBits, decoded.numHashes, decoded.seed, decoded.words, original.numBits, original.numHashes, original.seed, original.words)
	assertContainsMembers(t, decoded.ContainsString)
}

func TestAtomicFilterJSONRoundTrip(t *testing.T) {
	original := NewAtomicForSeeded(2000, 0.01, 42)
	for i := 0; i < 2000; i++ {
		original.AddString("member-" + strconv.Itoa(i))
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var decoded AtomicFilter
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	assertFilterState(t, decoded.numBits, decoded.numHashes, decoded.seed, snapshotAtomicWords(&decoded), original.numBits, original.numHashes, original.seed, snapshotAtomicWords(original))
	assertContainsMembers(t, decoded.ContainsString)
}

func TestAtomicFilterUnmarshalJSONRejectsInvalidData(t *testing.T) {
	var decoded AtomicFilter
	err := json.Unmarshal([]byte(`{"numBits":65,"numHashes":4,"seed":7,"words":[1]}`), &decoded)
	if err == nil {
		t.Fatalf("json.Unmarshal() unexpectedly succeeded for corrupt data")
	}
}

func TestAtomicFilterGobRoundTrip(t *testing.T) {
	original := NewAtomicForSeeded(2000, 0.01, 42)
	for i := 0; i < 2000; i++ {
		original.AddString("member-" + strconv.Itoa(i))
	}

	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(original); err != nil {
		t.Fatalf("gob Encode() error = %v", err)
	}

	var decoded AtomicFilter
	if err := gob.NewDecoder(&buf).Decode(&decoded); err != nil {
		t.Fatalf("gob Decode() error = %v", err)
	}

	assertFilterState(t, decoded.numBits, decoded.numHashes, decoded.seed, snapshotAtomicWords(&decoded), original.numBits, original.numHashes, original.seed, snapshotAtomicWords(original))
	assertContainsMembers(t, decoded.ContainsString)
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

	assertFilterState(t, decoded.numBits, decoded.numHashes, decoded.seed, snapshotAtomicWords(&decoded), original.numBits, original.numHashes, original.seed, snapshotAtomicWords(original))
	assertContainsMembers(t, decoded.ContainsString)
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

func assertFilterState(t *testing.T, gotBits uint64, gotHashes uint32, gotSeed uint64, gotWords []uint64, wantBits uint64, wantHashes uint32, wantSeed uint64, wantWords []uint64) {
	t.Helper()

	if gotBits != wantBits {
		t.Fatalf("numBits mismatch: got %d want %d", gotBits, wantBits)
	}
	if gotHashes != wantHashes {
		t.Fatalf("numHashes mismatch: got %d want %d", gotHashes, wantHashes)
	}
	if gotSeed != wantSeed {
		t.Fatalf("seed mismatch: got %d want %d", gotSeed, wantSeed)
	}
	if !equalWords(gotWords, wantWords) {
		t.Fatalf("word data mismatch after round trip")
	}
}

func assertContainsMembers(t *testing.T, contains func(string) bool) {
	t.Helper()

	for i := 0; i < 2000; i++ {
		value := "member-" + strconv.Itoa(i)
		if !contains(value) {
			t.Fatalf("missing inserted value %q after round trip", value)
		}
	}
}
