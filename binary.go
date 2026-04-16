package qbloom

import (
	"encoding/binary"
	"errors"
	"fmt"
	"sync/atomic"
)

const (
	filterBinaryMagic      = "QBLM"
	filterBinaryVersion    = uint32(1)
	filterBinaryHeaderSize = 4 + 4 + 8 + 4 + 8
)

// MarshalBinary implements encoding.BinaryMarshaler.
func (f *Filter) MarshalBinary() ([]byte, error) {
	if f == nil {
		return nil, errors.New("qbloom: cannot marshal nil Filter")
	}

	return marshalFilterBinary(f.words, f.numBits, f.numHashes, f.seed)
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler.
func (f *Filter) UnmarshalBinary(data []byte) error {
	if f == nil {
		return errors.New("qbloom: cannot unmarshal into nil Filter")
	}

	words, numBits, numHashes, seed, err := unmarshalFilterBinary(data)
	if err != nil {
		return err
	}

	f.words = words
	f.numBits = numBits
	f.numHashes = numHashes
	f.seed = seed
	return nil
}

// MarshalBinary implements encoding.BinaryMarshaler.
func (f *AtomicFilter) MarshalBinary() ([]byte, error) {
	if f == nil {
		return nil, errors.New("qbloom: cannot marshal nil AtomicFilter")
	}

	words := make([]uint64, len(f.words))
	for i := range f.words {
		words[i] = f.words[i].Load()
	}

	return marshalFilterBinary(words, f.numBits, f.numHashes, f.seed)
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler.
func (f *AtomicFilter) UnmarshalBinary(data []byte) error {
	if f == nil {
		return errors.New("qbloom: cannot unmarshal into nil AtomicFilter")
	}

	words, numBits, numHashes, seed, err := unmarshalFilterBinary(data)
	if err != nil {
		return err
	}

	atomicWords := make([]atomic.Uint64, len(words))
	for i, word := range words {
		atomicWords[i].Store(word)
	}

	f.words = atomicWords
	f.numBits = numBits
	f.numHashes = numHashes
	f.seed = seed
	return nil
}

func marshalFilterBinary(words []uint64, numBits uint64, numHashes uint32, seed uint64) ([]byte, error) {
	if numBits == 0 {
		return nil, errors.New("qbloom: cannot marshal filter with zero bits")
	}
	if numHashes == 0 {
		return nil, errors.New("qbloom: cannot marshal filter with zero hashes")
	}

	expectedWords := numBits / uint64(wordBits)
	if numBits%uint64(wordBits) != 0 || expectedWords != uint64(len(words)) {
		return nil, fmt.Errorf("qbloom: invalid filter shape: bits=%d words=%d", numBits, len(words))
	}

	data := make([]byte, filterBinaryHeaderSize+len(words)*8)
	copy(data[:4], filterBinaryMagic)
	binary.LittleEndian.PutUint32(data[4:8], filterBinaryVersion)
	binary.LittleEndian.PutUint64(data[8:16], numBits)
	binary.LittleEndian.PutUint32(data[16:20], numHashes)
	binary.LittleEndian.PutUint64(data[20:28], seed)

	offset := filterBinaryHeaderSize
	for _, word := range words {
		binary.LittleEndian.PutUint64(data[offset:offset+8], word)
		offset += 8
	}

	return data, nil
}

func unmarshalFilterBinary(data []byte) ([]uint64, uint64, uint32, uint64, error) {
	if len(data) < filterBinaryHeaderSize {
		return nil, 0, 0, 0, errors.New("qbloom: binary data too short")
	}
	if string(data[:4]) != filterBinaryMagic {
		return nil, 0, 0, 0, errors.New("qbloom: invalid binary header")
	}
	if version := binary.LittleEndian.Uint32(data[4:8]); version != filterBinaryVersion {
		return nil, 0, 0, 0, fmt.Errorf("qbloom: unsupported binary version %d", version)
	}

	numBits := binary.LittleEndian.Uint64(data[8:16])
	numHashes := binary.LittleEndian.Uint32(data[16:20])
	seed := binary.LittleEndian.Uint64(data[20:28])

	if numBits == 0 {
		return nil, 0, 0, 0, errors.New("qbloom: binary data has zero bits")
	}
	if numHashes == 0 {
		return nil, 0, 0, 0, errors.New("qbloom: binary data has zero hashes")
	}
	if numBits%uint64(wordBits) != 0 {
		return nil, 0, 0, 0, fmt.Errorf("qbloom: binary bit count %d is not word aligned", numBits)
	}

	wordData := data[filterBinaryHeaderSize:]
	if len(wordData)%8 != 0 {
		return nil, 0, 0, 0, errors.New("qbloom: binary word data is truncated")
	}

	expectedWords := numBits / uint64(wordBits)
	if expectedWords != uint64(len(wordData)/8) {
		return nil, 0, 0, 0, fmt.Errorf("qbloom: binary word count mismatch: bits=%d words=%d", numBits, len(wordData)/8)
	}

	words := make([]uint64, len(wordData)/8)
	for i := range words {
		offset := i * 8
		words[i] = binary.LittleEndian.Uint64(wordData[offset : offset+8])
	}

	return words, numBits, numHashes, seed, nil
}
