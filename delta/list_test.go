package delta

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestDeltaEntryIsEmpty(t *testing.T) {
	assert.Assert(t, Entry{Start: 0, End: 0, Source: SourceBasis}.isEmpty())
	assert.Assert(t, Entry{Start: 10, End: 9, Source: SourceBasis}.isEmpty())
	assert.Assert(t, Entry{Start: 10, End: 10, Source: SourceBasis}.isEmpty())
	assert.Assert(t, !Entry{Start: 10, End: 11, Source: SourceBasis}.isEmpty())
}

func TestAppendDelta(t *testing.T) {
	lst := []Entry{}
	lst = AppendDelta(lst, Entry{Start: 10, End: 10, Source: SourceBasis})
	assert.DeepEqual(t, lst, []Entry{})

	lst = AppendDelta(lst, Entry{Start: 0, End: 10, Source: SourceBasis})
	assert.DeepEqual(t, lst, []Entry{{Start: 0, End: 10, Source: SourceBasis}})

	lst = AppendDelta(lst, Entry{Start: 10, End: 20, Source: SourceBasis})
	assert.DeepEqual(t, lst, []Entry{{Start: 0, End: 20, Source: SourceBasis}})

	lst = AppendDelta(lst, Entry{Start: 20, End: 30, Source: SourceUpdate})
	assert.DeepEqual(t, lst, []Entry{
		{Start: 0, End: 20, Source: SourceBasis},
		{Start: 20, End: 30, Source: SourceUpdate},
	})
}
