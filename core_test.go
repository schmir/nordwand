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

var chunk_data []byte = make([]byte, 16384)

func init() {
	_, err := rand.Read(chunk_data)
	if err != nil {
		panic(err)
	}
}

type TestBuilder struct {
	basis, update *bytes.Buffer
	windowSize    int
	expected      []delta.DeltaEntry
}

func NewTestBuilder() *TestBuilder {
	tb := TestBuilder{
		basis:      &bytes.Buffer{},
		update:     &bytes.Buffer{},
		windowSize: 256,
		expected:   []delta.DeltaEntry{},
	}
	return &tb
}

func (tb *TestBuilder) Expect(ds ...delta.DeltaEntry) {
	for _, d := range ds {
		tb.expected = delta.AppendDelta(tb.expected, d)
	}
}

func (tb *TestBuilder) Run(t *testing.T) {
	sig := rsig.ComputeSignatureWithWindowSize(tb.basis.Bytes(), tb.windowSize)
	deltas := delta.ComputeDelta(sig, tb.update.Bytes())
	assert.DeepEqual(t, deltas, tb.expected)
}

func TestEmptyAppend(t *testing.T) {
	for _, size := range []int{0, 1, 2, 111, 888, 999} {
		t.Run(fmt.Sprintf("size=%d", size),
			func(t *testing.T) {
				tb := NewTestBuilder()
				_, _ = tb.update.Write(chunk_data[:size])
				tb.Expect(delta.DeltaEntry{Start: 0, End: uint64(size), Source: delta.SourceUpdate})
				tb.Run(t)
			},
		)
	}
}

func TestSameFile(t *testing.T) {
	for _, size := range []int{0, 1, 2, 111, 256, 888, 999, 2560, 3888} {
		t.Run(fmt.Sprintf("size=%d", size),
			func(t *testing.T) {
				tb := NewTestBuilder()
				tb.basis.Write(chunk_data[:size])
				tb.update.Write(chunk_data[:size])
				tb.Expect(delta.DeltaEntry{Start: 0, End: uint64(size), Source: delta.SourceBasis})
				tb.Run(t)
			},
		)
	}
}
