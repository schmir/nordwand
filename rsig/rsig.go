package rsig

import (
	"crypto/sha256"
	"hash/adler32"
)

const (
	minimumWindowSize = 128
	maximumWindowSize = 16384
)

type ChunkSignature struct {
	Sha256sum   [32]byte
	Adler32Hash uint32
}

type Signature struct {
	WindowSize uint32
	Chunks     []ChunkSignature
}

func (s *Signature) addChunk(data []byte) {
	s.Chunks = append(s.Chunks,
		ChunkSignature{
			Sha256sum:   sha256.Sum256(data),
			Adler32Hash: adler32.Checksum(data),
		},
	)
}

func chooseWindowSize(size int) int {
	// halve the window size until we have at least 512 chunks
	for windowSize := maximumWindowSize; windowSize >= minimumWindowSize; windowSize /= 2 {
		if size/windowSize >= 512 {
			return windowSize
		}
	}

	return minimumWindowSize
}

func ComputeSignatureWithWindowSize(basis []byte, size int) *Signature {
	signature := Signature{
		WindowSize: uint32(size),
	}

	for start := 0; start < len(basis); start += size {
		end := min(len(basis), start+size)
		signature.addChunk(basis[start:end])
	}
	return &signature
}

func ComputeSignature(basis []byte) *Signature {
	return ComputeSignatureWithWindowSize(
		basis,
		chooseWindowSize(len(basis)),
	)
}
