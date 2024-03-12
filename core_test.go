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

var chunkData = make([]byte, 16384)

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
}

func NewTestBuilder() *TestBuilder {
	tb := TestBuilder{
		basis:      &bytes.Buffer{},
		update:     &bytes.Buffer{},
		windowSize: 256,
		expected:   []delta.Entry{},
	}
	return &tb
}

func (tb *TestBuilder) Expect(ds ...delta.Entry) {
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
				_, _ = tb.update.Write(chunkData[:size])
				tb.Expect(delta.Entry{Start: 0, End: uint64(size), Source: delta.SourceUpdate})
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
				tb.basis.Write(chunkData[:size])
				tb.update.Write(chunkData[:size])
				tb.Expect(delta.Entry{Start: 0, End: uint64(size), Source: delta.SourceBasis})
				tb.Run(t)
			},
		)
	}
}
