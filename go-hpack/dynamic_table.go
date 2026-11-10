package hpack

type dynamicTable struct {
	entries []HeaderField
	size    int
	maxSize int
}

func newDynamicTable(maxSize int) *dynamicTable {
	return &dynamicTable{
		entries: make([]HeaderField, 0, 16),
		maxSize: maxSize,
	}
}

func (dt *dynamicTable) add(hf HeaderField) {
	entrySize := hf.Size()
	if entrySize > dt.maxSize {
		// Clear table if entry exceeds maximum size alone
		dt.entries = dt.entries[:0]
		dt.size = 0
		return
	}

	// Evict from back until fits
	for dt.size+entrySize > dt.maxSize && len(dt.entries) > 0 {
		lastIdx := len(dt.entries) - 1
		dt.size -= dt.entries[lastIdx].Size()
		dt.entries = dt.entries[:lastIdx]
	}

	// Prepend entry (newest at index 0)
	dt.entries = append([]HeaderField{hf}, dt.entries...)
	dt.size += entrySize
}

func (dt *dynamicTable) get(index int) (HeaderField, bool) {
	if index < 1 || index > len(dt.entries) {
		return HeaderField{}, false
	}
	return dt.entries[index-1], true
}

func (dt *dynamicTable) setMaxSize(newMax int) {
	dt.maxSize = newMax
	for dt.size > dt.maxSize && len(dt.entries) > 0 {
		lastIdx := len(dt.entries) - 1
		dt.size -= dt.entries[lastIdx].Size()
		dt.entries = dt.entries[:lastIdx]
	}
}
