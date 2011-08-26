package dgobloom

import (
	"hash"
	"math"
)

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

type BloomFilter struct {
	capacity uint32
	elements uint32
	bits     uint32   // size of bit vector in bits
	filter   []uint32 // our filter bit vector
	h        hash.Hash32
	salts    [][]byte
}

func (bf *BloomFilter) Elements() uint32 { return bf.elements }

func FilterBits(capacity uint32, falsePositiveRate float64) uint32 {
	bits := float64(capacity) * -math.Log(falsePositiveRate) / (math.Log(2.0) * math.Log(2.0)) // in bits
	m := nextPowerOfTwo(uint32(bits))

	if m < 1024 {
		return 1024
	}

	return m
}

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

func NewBloomFilter(capacity uint32, falsePositiveRate float64, h hash.Hash32, salts []uint32) *BloomFilter {

	bf := new(BloomFilter)

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

func (bf *BloomFilter) Insert(b []byte) bool {

	bf.elements++

	for _, s := range bf.salts {
		bf.h.Reset()
		bf.h.Write(s)
		bf.h.Write(b)
		setbit(bf.filter, bf.h.Sum32()%bf.bits)
	}

	return bf.elements < bf.capacity
}

func (bf *BloomFilter) Exists(b []byte) bool {

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

func (bf *BloomFilter) FillPercentage() float64 {

	set := uint(0)
	for i := uint32(0); i < bf.bits; i++ {
		set += getbit(bf.filter, i)
	}

	return float64(set) / float64(bf.bits)
}
