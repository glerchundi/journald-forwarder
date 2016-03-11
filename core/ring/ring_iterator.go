package ring

import "github.com/glerchundi/go-systemd/sdjournal"

type RingIterator struct {
	ring     *Ring
	index    int
	current  *sdjournal.JournalEntry
	finished bool
}

func (i *RingIterator) Reset() {
	i.index = 0
	i.current = nil
	i.finished = false
}

func (i *RingIterator) Next() bool {
	if i.finished {
		return false
	}

	if i.ring.head == -1 {
		return false
	}

	idx := i.ring.mod(i.index  + i.ring.tail)
	i.current = i.ring.get(idx)
	if idx == i.ring.head {
		i.finished = true
	}

	i.index++
	return true
}

func (i *RingIterator) Value() (int, *sdjournal.JournalEntry) {
	return i.index, i.current
}