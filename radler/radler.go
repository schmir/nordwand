// package radler implements a rolling adler32 checksum
package radler

const (
	// mod is the largest prime that is less than 65536.
	mod = 65521
)

// Radler implements a rolling adler32 checksum. This has been optimized for readability, not
// speed.
type Radler struct {
	s1, s2 uint32
	window []byte
	pos    uint32
	full   bool
}

// Checksum returns the current checksum
func (r *Radler) Checksum() uint32 {
	return (r.s2 << 16) | r.s1
}

// pushOut a single byte from the checksum
func (r *Radler) pushOut(outgoing uint32) {
	// substract once from s1
	negOutgoing := mod - outgoing
	r.s1 = (r.s1 + negOutgoing) % mod

	// substract window size times from s2
	windowSize := uint32((len(r.window)) % mod)
	r.s2 = (r.s2 + negOutgoing*windowSize - 1) % mod
}

// Push adds a byte and if the window is full, also removes the oldest byte from the window
func (r *Radler) Push(b byte) {
	if r.full {
		r.pushOut(uint32(r.window[r.pos]))
	}

	incoming := uint32(b)
	r.s1 = (r.s1 + incoming) % mod
	r.s2 = (r.s2 + r.s1) % mod

	r.window[r.pos] = b
	r.pos += 1
	if r.pos >= uint32(len(r.window)) {
		r.pos = 0
		r.full = true
	}
}

// New creates a new Radler struct.
func New(windowSize uint32) *Radler {
	return &Radler{
		s1:     1,
		s2:     0,
		window: make([]byte, windowSize),
		pos:    0,
		full:   false,
	}
}
