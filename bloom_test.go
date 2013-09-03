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
)

const CAPACITY = 10000
const ERRPCT = 0.01

func TestBloomFilter(t *testing.T) {

	saltsNeeded := SaltsRequired(CAPACITY, ERRPCT)

	t.Log("generating", saltsNeeded, "salts")

	salts := make([]uint32, saltsNeeded)

	for i := uint(0); i < saltsNeeded; i++ {
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

	errorPct := errors / total

	t.Log("error percentage: (", errors, "/", total, ")=", errorPct)

	if errorPct > ERRPCT {
		t.Fail()
	}
}
