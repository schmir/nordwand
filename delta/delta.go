package delta

import (
	"fmt"

	"github.com/schmir/nordwand/radler"
	"github.com/schmir/nordwand/rsig"
)

type deltaBuilder struct {
	signature                      *rsig.Signature
	chksum                         *radler.Radler
	chunkChecksums                 map[uint32]struct{}
	secureHashMap                  map[[32]byte]int
	deltaEntryStart, deltaEntryEnd uint64
	delta                          []DeltaEntry
}

func newDeltaBuilder(basisSignature *rsig.Signature) *deltaBuilder {
	db := &deltaBuilder{
		signature:       basisSignature,
		chksum:          radler.New(basisSignature.WindowSize),
		chunkChecksums:  make(map[uint32]struct{}),
		secureHashMap:   make(map[[32]byte]int),
		deltaEntryStart: 0,
		deltaEntryEnd:   0,
		delta:           []DeltaEntry{},
	}
	for i, chunk := range basisSignature.Chunks {
		db.chunkChecksums[chunk.Adler32Hash] = struct{}{}
		db.secureHashMap[chunk.Sha256sum] = i
	}
	return db
}

func (db *deltaBuilder) findChunk() (int, bool) {
	_, ok := db.chunkChecksums[db.chksum.Checksum()]
	if !ok {
		return 0, false
	}
	// the rolling checksum matches, let's see if we have a match when using the secure hash
	i, ok := db.secureHashMap[db.chksum.Sum256()]
	if !ok {
		return 0, false
	}

	fmt.Printf("found a chunk: %d\n", i)
	return i, true
}

func (db *deltaBuilder) storeChunkFound(i int) {
	db.delta = AppendDelta(db.delta, DeltaEntry{
		Start:  db.deltaEntryStart,
		End:    db.deltaEntryEnd - uint64(db.signature.WindowSize),
		Source: SourceUpdate,
	})
	db.delta = AppendDelta(db.delta, DeltaEntry{
		Start:  uint64(i) * uint64(db.signature.WindowSize),
		End:    uint64(i)*uint64(db.signature.WindowSize) + uint64(db.chksum.Size()),
		Source: SourceBasis,
	})

	db.deltaEntryStart = db.deltaEntryEnd
	db.chksum.Reset()
}

func (db *deltaBuilder) push(b byte) {
	db.chksum.Push(b)
	db.deltaEntryEnd += 1
	if !db.chksum.IsFull() {
		return
	}
	i, found := db.findChunk()
	if !found {
		return
	}
	db.storeChunkFound(i)
}

func (db *deltaBuilder) close() {
	i, found := db.findChunk()
	if found {
		db.storeChunkFound(i)
	} else {
		db.delta = AppendDelta(db.delta, DeltaEntry{
			Start:  db.deltaEntryStart,
			End:    db.deltaEntryEnd,
			Source: SourceUpdate,
		})
		db.deltaEntryStart = db.deltaEntryEnd
		db.chksum.Reset()
	}
}

func (db *deltaBuilder) Write(data []byte) (n int, err error) {
	for _, b := range data {
		db.push(b)
	}

	return len(data), nil
}

func ComputeDelta(basisSignature *rsig.Signature, updated []byte) []DeltaEntry {
	builder := newDeltaBuilder(basisSignature)
	_, _ = builder.Write(updated)
	builder.close()
	return builder.delta
}
