package main

import (
	"fmt"

	"github.com/schmir/nordwand/delta"
	"github.com/schmir/nordwand/rsig"
)

func main() {
	data := [256]byte{}
	for i := 0; i < 256; i++ {
		data[i] = byte(i)
	}
	basis := data[:]
	signature := rsig.ComputeSignatureWithWindowSize(basis, 23)
	delta := delta.ComputeDelta(signature, basis)
	fmt.Printf("DELTA: %v", delta)
}
