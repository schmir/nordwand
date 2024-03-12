package main_test

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"testing"

	"gotest.tools/v3/assert"

	"github.com/schmir/nordwand/delta"
	"github.com/schmir/nordwand/rsig"
)

var chunkData = make([]byte, 1024*1024)

func init() {
	_, err := rand.Read(chunkData)
	if err != nil {
		panic(err)
	}
}

type TestBuilder struct {
	basis, update *bytes.Buffer
	windowSize    int
	expected      []delta.Entry
	numRuns       int
}

func NewTestBuilder() *TestBuilder {
	tb := TestBuilder{
		basis:      &bytes.Buffer{},
		update:     &bytes.Buffer{},
		windowSize: 100,
		expected:   []delta.Entry{},
	}
	return &tb
}

func (tb *TestBuilder) Chunk(n int) []byte {
	return chunkData[n*tb.windowSize : (n+1)*tb.windowSize]
}

func (tb *TestBuilder) Expect(ds ...delta.Entry) {
	for _, d := range ds {
		tb.expected = delta.AppendDelta(tb.expected, d)
	}
}

// InsertRand inserts num random bytes into the updated file.
func (tb *TestBuilder) InsertRand(t *testing.T, num int) {
	d := make([]byte, num)
	_, _ = rand.Read(d)
	start := tb.update.Len()
	tb.update.Write(d)
	tb.Expect(delta.Entry{
		Start:  uint64(start),
		End:    uint64(tb.update.Len()),
		Source: delta.SourceUpdate,
	})
	tb.Run(t)
}

// InsertChunk inserts the given chunk into the updated file and updates the expected result. The
// given chunk must be part of the basis.
func (tb *TestBuilder) InsertChunk(t *testing.T, n int) {
	tb.update.Write(tb.Chunk(n))
	tb.Expect(delta.Entry{
		Start:  uint64(n * tb.windowSize),
		End:    uint64((n + 1) * tb.windowSize),
		Source: delta.SourceBasis,
	})
	tb.Run(t)
}

func (tb *TestBuilder) ResetUpdate() {
	tb.expected = []delta.Entry{}
	tb.update = &bytes.Buffer{}
}

func (tb *TestBuilder) Run(t *testing.T) {
	tb.numRuns++
	t.Run(fmt.Sprintf("%d basis:%d updated: %d", tb.numRuns, len(tb.basis.Bytes()), len(tb.update.Bytes())),
		func(t *testing.T) {
			assert.Equal(t, delta.FileSize(tb.expected), uint64(len(tb.update.Bytes())))

			sig := rsig.ComputeSignatureWithWindowSize(tb.basis.Bytes(), tb.windowSize)
			deltas := delta.ComputeDelta(sig, tb.update.Bytes())
			fmt.Printf("%3d:  have: %v\n  expected: %v\n", tb.numRuns, deltas, tb.expected)
			assert.DeepEqual(t, deltas, tb.expected)
		},
	)
}

func TestEmptyAppend(t *testing.T) {
	for _, size := range []int{0, 1, 2, 111, 888, 999} {
		tb := NewTestBuilder()
		_, _ = tb.update.Write(chunkData[:size])
		tb.Expect(delta.Entry{Start: 0, End: uint64(size), Source: delta.SourceUpdate})
		tb.Run(t)
	}
}

func TestSameFile(t *testing.T) {
	for _, size := range []int{0, 1, 2, 111, 256, 888, 999, 2560, 3888} {
		tb := NewTestBuilder()
		tb.basis.Write(chunkData[:size])
		tb.update.Write(chunkData[:size])
		tb.Expect(delta.Entry{Start: 0, End: uint64(size), Source: delta.SourceBasis})
		tb.Run(t)
	}
}

func TestComplex(t *testing.T) {
	tb := NewTestBuilder()
	for i := range 20 {
		tb.basis.Write(tb.Chunk(i))
	}
	tb.Run(t)
	tb.InsertRand(t, 99)
	tb.InsertChunk(t, 0)
	tb.InsertChunk(t, 3)

	tb.InsertRand(t, 456457)

	tb.InsertChunk(t, 5)
	tb.InsertRand(t, 453)
	tb.InsertChunk(t, 5)
	tb.InsertChunk(t, 3)

	tb.Run(t)
}

// TestTail demonstrates what happens with the tail chunk, i.e. the bytes at the end of the input
// that do not form a complete chunk. When we build the signature, we do store a ChunkSignature at
// the end. This allows us to find the tail in some special cases. I though about removing the
// signature for tail, but then we cannot detect unchanged files.
func TestTail(t *testing.T) {
	tail := []byte{20, 19, 18, 17, 16, 15, 14, 13}
	tb := NewTestBuilder()
	for i := range 20 {
		tb.basis.Write(tb.Chunk(i))
	}
	basisTailStart := uint64(len(tb.basis.Bytes()))
	tb.basis.Write(tail)
	basisTailEnd := uint64(len(tb.basis.Bytes()))

	ensureTailFound := func() {
		tb.Expect(delta.Entry{
			Start:  basisTailStart,
			End:    basisTailEnd,
			Source: delta.SourceBasis,
		})
		tb.Run(t)
	}

	// Here we can find the tail because there's nothing in front
	tb.update.Write(tail)
	ensureTailFound()

	// Here we can find the tail because we have a chunk directly in front of it
	tb.ResetUpdate()
	tb.InsertRand(t, 68)
	tb.InsertChunk(t, 5)
	tb.update.Write(tail)
	ensureTailFound()

	// Here we cannot find the tail because we have some non-chunk data in front of it
	tb.ResetUpdate()
	tb.InsertRand(t, 2)
	start := uint64(len(tb.update.Bytes()))
	tb.update.Write(tail)
	tb.Expect(delta.Entry{
		Start:  start,
		End:    start + uint64(len(tail)),
		Source: delta.SourceUpdate,
	})

	// Cannot find the tail here because the tail doesn't start at a chunk boundary
	tb.ResetUpdate()
	tb.update.Write([]byte{1})
	tb.update.Write(tail)
	tb.Expect(delta.Entry{
		Start:  0,
		End:    uint64(1 + len(tail)),
		Source: delta.SourceUpdate,
	},
	)
	tb.Run(t)

	// Cannot find the tail here because the tail doesn't start at a chunk boundary
	tb.ResetUpdate()
	tb.update.Write(tail)
	tb.update.Write([]byte{1})
	tb.Expect(delta.Entry{
		Start:  0,
		End:    uint64(1 + len(tail)),
		Source: delta.SourceUpdate,
	},
	)
	tb.Run(t)
}
