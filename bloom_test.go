// Unit tests for the bloom filter
// Copyright (c) 2011 Damian Gryski <damian@gryski.com>
// Licensed under the GPLv3, or at your option any later version.
package dgobloom

import (
	"bufio"
	"hash/fnv"
	"math/rand"
	"os"
	"testing"
	//        "../dgohash.git/_obj/dgohash"
)

const CAPACITY = 100000
const ERRPCT = 0.01

func TestBloomFilter(t *testing.T) {

	salts_needed := SaltsRequired(CAPACITY, ERRPCT)

	t.Log("generating", salts_needed, "salts")

	salts := make([]uint32, salts_needed)

	for i := uint(0); i < salts_needed; i++ {
		salts[i] = rand.Uint32()
	}

	b := NewBloomFilter(CAPACITY, ERRPCT, fnv.New32(), salts)

	fh, _ := os.Open("/usr/share/dict/words")

	buf := bufio.NewReader(fh)

	for {
		l, _, err := buf.ReadLine()
		if err != nil {
			break
		}

		if !b.Insert(l) {
			break
		}
	}

	t.Log("inserted", b.Elements(), "elements")

	total := 0.0
	errors := 0.0

	for {
		l, _, err := buf.ReadLine()
		if err != nil {
			break
		}

		if b.Exists(l) {
			errors++
		}
		total++
	}

	error_pct := errors / total

	t.Log("error percentage: (", errors, "/", total, ")=", error_pct)

	if error_pct > ERRPCT {
		t.Fail()
	}
}
