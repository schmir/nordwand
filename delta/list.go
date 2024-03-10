package delta

type DeltaSource int8

const (
	_                       = iota
	SourceBasis DeltaSource = iota
	SourceUpdate
)

// DeltaEntry refers to a range inside either the original or the new file
type DeltaEntry struct {
	Start, End uint64
	Source     DeltaSource
}

// mergeDeltaEntry merges the next DeltaEntry with this one if possible. It return true on success.
func (entry *DeltaEntry) mergeDeltaEntry(next DeltaEntry) bool {
	if entry.End != next.Start || entry.Source != next.Source {
		return false
	}
	entry.End = next.End
	return true
}

func (entry DeltaEntry) isEmpty() bool {
	return entry.End <= entry.Start
}

// AppendDelta appends the given DeltaEntry to the list. It will ignore empty DeltaEntries and
// merge the given DeltaEntry with the last one if possible.
func AppendDelta(lst []DeltaEntry, delta DeltaEntry) []DeltaEntry {
	if delta.isEmpty() {
		return lst
	}
	if len(lst) > 0 && lst[len(lst)-1].mergeDeltaEntry(delta) {
		return lst
	}
	return append(lst, delta)
}
