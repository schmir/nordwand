package radler

import (
	"crypto/rand"
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

// TestRadler tests the rolling adler implementation by comparing against the stdlib's adler32 package
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
