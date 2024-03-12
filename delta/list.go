package delta

type Source int8

const (
	_                  = iota
	SourceBasis Source = iota
	SourceUpdate
)

// Entry refers to a range inside either the original or the new file.
type Entry struct {
	Start, End uint64
	Source     Source
}

// mergeDeltaEntry merges the next DeltaEntry with this one if possible. It return true on success.
func (entry *Entry) mergeDeltaEntry(next Entry) bool {
	if entry.End != next.Start || entry.Source != next.Source {
		return false
	}
	entry.End = next.End
	return true
}

func (entry Entry) isEmpty() bool {
	return entry.End <= entry.Start
}

// AppendDelta appends the given DeltaEntry to the list. It will ignore empty DeltaEntries and
// merge the given DeltaEntry with the last one if possible.
func AppendDelta(lst []Entry, delta Entry) []Entry {
	if delta.isEmpty() {
		return lst
	}
	if len(lst) > 0 && lst[len(lst)-1].mergeDeltaEntry(delta) {
		return lst
	}
	return append(lst, delta)
}
