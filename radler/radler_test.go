package radler

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"hash/adler32"
	"testing"

	"gotest.tools/v3/assert"
)

type chksum struct {
	s1, s2 uint32
}

func newChksum(s uint32) chksum {
	return chksum{
		s1: s & 0xffff,
		s2: s >> 16,
	}
}

// TestRadler tests the rolling adler implementation by comparing against the stdlib's adler32 package.
func TestRadler(t *testing.T) {
	data := make([]byte, 10*1024*1024)
	_, err := rand.Read(data)
	if err != nil {
		panic(err)
	}

	for _, windowSize := range []uint32{1, 2, 7, 128, 512, 2046, 8192, 128 * 1024} {
		t.Run(fmt.Sprintf("%d", windowSize),
			func(t *testing.T) {
				r := New(windowSize)
				for i, b := range data[:windowSize+512] {
					r.Push(b)
					end := i + 1
					start := max(0, end-int(windowSize))
					expected := newChksum(adler32.Checksum(data[start:end]))
					got := newChksum(r.Checksum())
					assert.Equal(t, got, expected, "%d %d %t\n", start, end, r.full)
				}
			},
		)
	}
}

// TestRadlerSum256 tests that Sum256 works when the window is empty, half-filled, full, and when rolled.
func TestRadlerSum256(t *testing.T) {
	windowSize := 16
	data := make([]byte, 3*windowSize)
	for i := range len(data) {
		data[i] = byte(i)
	}

	r := New(uint32(windowSize))
	for i, b := range data {
		assert.Equal(t, r.Sum256(), sha256.Sum256(data[max(0, i-windowSize):i]))
		r.Push(b)
	}
}
