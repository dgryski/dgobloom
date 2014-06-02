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
	b2 := NewBloomFilter(CAPACITY, ERRPCT, fnv.New32(), salts)

	fh, _ := os.Open("/usr/share/dict/words")

	buf := bufio.NewReader(fh)

	for {
		l, _, err := buf.ReadLine()
		if err != nil {
			break
		}

		b2.Insert(l)

		if !b.Insert(l) {
			break
		}
	}

	t.Log("inserted", b.Elements(), "elements")

	total := 0.0
	errors := 0.0
	errors2 := 0.0

	b2.Compress()

	for {
		l, _, err := buf.ReadLine()
		if err != nil {
			break
		}

		if b.Exists(l) {
			errors++
		}

		if b2.Exists(l) {
			errors2++
		}

		total++
	}

	errorPct := errors / total

	t.Log("error percentage: (", errors, "/", total, ")=", errorPct)
	t.Log("error percentage2: (", errors2, "/", total, ")=", errors2/total)

	if errorPct > ERRPCT {
		t.Fail()
	}
}
