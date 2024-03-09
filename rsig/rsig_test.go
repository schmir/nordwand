package rsig

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestChooseWindowSize(t *testing.T) {
	assert.Equal(t, chooseWindowSize(0), minimumWindowSize)

	assert.Equal(t, chooseWindowSize(512*2*minimumWindowSize-1), minimumWindowSize)
	assert.Equal(t, chooseWindowSize(512*2*minimumWindowSize), 2*minimumWindowSize)

	assert.Equal(t, chooseWindowSize(512*maximumWindowSize-1), maximumWindowSize/2)
	assert.Equal(t, chooseWindowSize(512*maximumWindowSize), maximumWindowSize)

	assert.Equal(t, chooseWindowSize(8*512*maximumWindowSize), maximumWindowSize)
}

func TestComputeSignatureEmpty(t *testing.T) {
	sig := ComputeSignatureWithWindowSize([]byte{}, 128)
	assert.DeepEqual(t, sig, &Signature{WindowSize: 128})
}

// XXX needs more tests
