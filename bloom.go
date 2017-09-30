package bloom

import (
	"bytes"
	"errors"

	"github.com/spaolacci/murmur3"
	"github.com/willf/bitset"
)

// A BloomFilter is a representation of a set of _n_ items, where the main
// requirement is to make membership queries; _i.e._, whether an item is a
// member of a set.
type BloomFilter struct {
	m uint
	k uint
	b *bitset.BitSet
}

func max(x, y uint) uint {
	if x > y {
		return x
	}
	return y
}

func uintToBytes(value uint) []byte {
	b := make([]byte, 4)
	b[0] = (byte)(value)
	b[1] = (byte)(value >> 8)
	b[2] = (byte)(value >> 16)
	b[3] = (byte)(value >> 24)

	return b
}

// New creates a new Bloom filter with _m_ bits and _k_ hashing functions
// We force _m_ and _k_ to be at least one to avoid panics.
func New(m uint, k uint) (*BloomFilter, error) {
	if m < 1 {
		return nil, errors.New("New m < 1")
	} else if k < 1 {
		return nil, errors.New("New k < 1")
	} else {
		return &BloomFilter{m, k, bitset.New(m)}, nil
	}
}

// location returns the ith hashed location using the four base hash values
func (f *BloomFilter) location(data []byte, i uint) uint {
	tempData := [][]byte{data, uintToBytes(i)}
	result := bytes.Join(tempData, []byte(""))
	hasher := murmur3.New64()
	hasher.Write(result)
	location := uint(hasher.Sum64()) % f.m

	return location
}

// Cap returns the capacity, _m_, of a Bloom filter
func (f *BloomFilter) Cap() uint {
	return f.m
}

// K returns the number of hash functions used in the BloomFilter
func (f *BloomFilter) HashFunctionNum() uint {
	return f.k
}

// Add data to the Bloom Filter. Returns the filter (allows chaining)
func (f *BloomFilter) Add(data []byte) *BloomFilter {
	for i := uint(0); i < f.k; i++ {
		f.b.Set(f.location(data, i))
	}
	return f
}

// AddString to the Bloom Filter. Returns the filter (allows chaining)
func (f *BloomFilter) AddString(data string) *BloomFilter {
	return f.Add([]byte(data))
}

// Test returns true if the data is in the BloomFilter, false otherwise.
// If true, the result might be a false positive. If false, the data
// is definitely not in the set.
func (f *BloomFilter) Test(data []byte) bool {
	for i := uint(0); i < f.k; i++ {
		if !f.b.Test(f.location(data, i)) {
			return false
		}
	}
	return true
}

// TestString returns true if the string is in the BloomFilter, false otherwise.
// If true, the result might be a false positive. If false, the data
// is definitely not in the set.
func (f *BloomFilter) TestString(data string) bool {
	return f.Test([]byte(data))
}

// TestLocations returns true if all locations are set in the BloomFilter, false
// otherwise.
func (f *BloomFilter) TestLocations(locs []uint64) bool {
	for i := 0; i < len(locs); i++ {
		if !f.b.Test(uint(locs[i] % uint64(f.m))) {
			return false
		}
	}
	return true
}

// TestAndAdd is the equivalent to calling Test(data) then Add(data).
// Returns the result of Test.
func (f *BloomFilter) TestAndAdd(data []byte) bool {
	present := true
	for i := uint(0); i < f.k; i++ {
		l := f.location(data, i)
		if !f.b.Test(l) {
			present = false
		}
		f.b.Set(l)
	}
	return present
}

// TestAndAddString is the equivalent to calling Test(string) then Add(string).
// Returns the result of Test.
func (f *BloomFilter) TestAndAddString(data string) bool {
	return f.TestAndAdd([]byte(data))
}

// ClearAll clears all the data in a Bloom filter, removing all keys
func (f *BloomFilter) ClearAll() *BloomFilter {
	f.b.ClearAll()
	return f
}
