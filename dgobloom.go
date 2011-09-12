/* Package dgobloom implements a simple Bloom Filter for strings.

   A Bloom Filter is a probablistic data structure which allows testing set membership.
   A negative answer means the value is not in the set.  A positive answer means the element
   is probably is the set.  The desired rate false positives can be set at filter construction time.

   Copyright (c) 2011 Damian Gryski <damian@gryski.com>

   Licensed under the GPLv3, or at your option any later version.
*/

package dgobloom

import (
	"hash"
	"math"
)

// Internal routines for the bit vector
func getbit(d []uint32, bit uint32) uint {

	shift := bit % 32
	bb := d[bit/32]
	bb &= (1 << shift)

	return uint(bb >> shift)
}

func setbit(d []uint32, bit uint32) {
	d[bit/32] |= (1 << (bit % 32))
}

func clearbit(d []uint32, bit uint32) {
	d[bit/32] &= ^(1 << (bit % 32))
}

// 32-bit, which is why it only goes up to 16
// return the integer >= i which is a power of two
func nextPowerOfTwo(i uint32) uint32 {
	n := i - 1
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	n++
	return n
}

type BloomFilter interface {
	// Insert an element into the set.
	Insert(b []byte) bool

	// Determine if an element is in the set
	Exists(b []byte) bool

	// Return the number of elements currently stored in the set
	Elements() uint32
}

// Internal struct for our bloom filter
type bloomFilter struct {
	capacity uint32
	elements uint32
	bits     uint32   // size of bit vector in bits
	filter   []uint32 // our filter bit vector
	h        hash.Hash32
	salts    [][]byte
}

func (bf *bloomFilter) Elements() uint32 { return bf.elements }

// FilterBits returns the number of bits required for the desired capacity and false positive rate.
func FilterBits(capacity uint32, falsePositiveRate float64) uint32 {
	bits := float64(capacity) * -math.Log(falsePositiveRate) / (math.Log(2.0) * math.Log(2.0)) // in bits
	m := nextPowerOfTwo(uint32(bits))

	if m < 1024 {
		return 1024
	}

	return m
}

// SaltsRequired returns the number of salts required by the constructor for the desired capacity and false positive rate.
func SaltsRequired(capacity uint32, falsePositiveRate float64) uint {
	m := FilterBits(capacity, falsePositiveRate)
	salts := uint(0.7 * float32(float64(m)/float64(capacity)))
	if salts < 2 {
		return 2
	}
	return salts
}

func uint32ToByteArray(salt uint32) []byte {
	p := make([]byte, 4)
	p[0] = byte(salt >> 24)
	p[1] = byte(salt >> 16)
	p[2] = byte(salt >> 8)
	p[3] = byte(salt)
	return p
}

// NewBloomFilter returns a new bloom filter with the specified capacity and false positive rate.
// The hash function h will be salted with the array of salts.
func NewBloomFilter(capacity uint32, falsePositiveRate float64, h hash.Hash32, salts []uint32) BloomFilter {

	bf := new(bloomFilter)

	bf.capacity = capacity
	bf.bits = FilterBits(capacity, falsePositiveRate)
	bf.filter = make([]uint32, uint(bf.bits+31)/32)
	bf.h = h

	bf.salts = make([][]byte, len(salts))
	for i, s := range salts {
		bf.salts[i] = uint32ToByteArray(s)
	}

	return bf
}

// Insert inserts the byte array b into the bloom filter.
// If the function returns false, the capacity of the bloom filter has been reached.  Further inserts will increase the rate of false positives.
func (bf *bloomFilter) Insert(b []byte) bool {

	bf.elements++

	for _, s := range bf.salts {
		bf.h.Reset()
		bf.h.Write(s)
		bf.h.Write(b)
		setbit(bf.filter, bf.h.Sum32()%bf.bits)
	}

	return bf.elements < bf.capacity
}

// Exists checks the bloom filter for the byte array b
func (bf *bloomFilter) Exists(b []byte) bool {

	for _, s := range bf.salts {
		bf.h.Reset()
		bf.h.Write(s)
		bf.h.Write(b)

		if getbit(bf.filter, bf.h.Sum32()%bf.bits) == 0 {
			return false
		}
	}

	return true

}
