/*
Package ring provides a simple implementation of a ring buffer.
*/
package ring

import "github.com/glerchundi/go-systemd/sdjournal"

/*
The DefaultCapacity of an uninitialized Ring buffer.

Changing this value only affects ring buffers created after it is changed.
*/
var DefaultCapacity int = 10

/*
Type Ring implements a Circular Buffer.
The default value of the Ring struct is a valid (empty) Ring buffer with capacity DefaultCapacify.
*/
type Ring struct {
	head     int // the most recent value written
	tail     int // the least recent value written
	buff     []*sdjournal.JournalEntry
	iterator *RingIterator
}

func NewRing(size int) (ring *Ring) {
	ring = &Ring{}
	ring.SetCapacity(size)
	return
}

/*
Set the maximum size of the ring buffer.
*/
func (r *Ring) SetCapacity(size int) {
	r.checkInit()
	r.extend(size)
}

/*
Capacity returns the current capacity of the ring buffer.
*/
func (r Ring) Capacity() int {
	return len(r.buff)
}

/*
Enqueue a value into the Ring buffer.
*/
func (r *Ring) Enqueue(i *sdjournal.JournalEntry) {
	r.checkInit()
	r.set(r.head+1, i)
	old := r.head
	r.head = r.mod(r.head + 1)
	if old != -1 && r.head == r.tail {
		r.tail = r.mod(r.tail + 1)
	}
}

/*
Dequeue a value from the Ring buffer.

Returns nil if the ring buffer is empty.
*/
func (r *Ring) Dequeue() *sdjournal.JournalEntry {
	r.checkInit()
	if r.head == -1 {
		return nil
	}
	v := r.get(r.tail)
	if r.tail == r.head {
		r.head = -1
		r.tail = 0
	} else {
		r.tail = r.mod(r.tail + 1)
	}
	return v
}

/*
Read the value that Dequeue would have dequeued without actually dequeuing it.

Returns nil if the ring buffer is empty.
*/
func (r *Ring) Peek() *sdjournal.JournalEntry {
	r.checkInit()
	if r.head == -1 {
		return nil
	}
	return r.get(r.tail)
}

/*
Values returns a slice of all the values in the circular buffer without modifying them at all.
The returned slice can be modified independently of the circular buffer. However, the values inside the slice
are shared between the slice and circular buffer.
*/
func (r *Ring) Values() []*sdjournal.JournalEntry {
	if r.head == -1 {
		return []*sdjournal.JournalEntry{}
	}
	arr := make([]*sdjournal.JournalEntry, 0, r.Capacity())
	for i := 0; i < r.Capacity(); i++ {
		idx := r.mod(i + r.tail)
		arr = append(arr, r.get(idx))
		if idx == r.head {
			break
		}
	}
	return arr
}

/*
Returns the length of the used part of the buffer
*/
func (r *Ring) Len() int {
	if r.Capacity() == 0 {
		return 0
	}
	if r.head == -1 {
		return 0
	}
	if r.tail > r.head {
		return r.head - r.tail + r.Capacity() + 1
	} else {
		return r.head - r.tail + 1
	}
}

/**
*** Unexported methods beyond this point.
**/

// sets a value at the given unmodified index and returns the modified index of the value
func (r *Ring) set(p int, v *sdjournal.JournalEntry) {
	r.buff[r.mod(p)] = v
}

// gets a value based at a given unmodified index
func (r *Ring) get(p int) *sdjournal.JournalEntry {
	return r.buff[r.mod(p)]
}

// returns the modified index of an unmodified index
func (r *Ring) mod(p int) int {
	return p % len(r.buff)
}

func (r *Ring) checkInit() {
	if r.buff == nil {
		r.buff = make([]*sdjournal.JournalEntry, DefaultCapacity)
		for i := range r.buff {
			r.buff[i] = nil
		}
		r.head, r.tail = -1, 0
	}
}

func (r *Ring) extend(size int) {
	if size == len(r.buff) {
		return
	} else if size < len(r.buff) {
		r.buff = r.buff[0:size]
	}
	newb := make([]*sdjournal.JournalEntry, size-len(r.buff))
	for i := range newb {
		newb[i] = nil
	}
	r.buff = append(r.buff, newb...)
}

func (r *Ring) Iterator() (iterator *RingIterator) {
	if r.iterator != nil {
		r.iterator.Reset()
	} else {
		r.iterator = &RingIterator{r, 0, nil, false}
	}
	return r.iterator
}
