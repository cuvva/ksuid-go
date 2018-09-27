package ksuid

import (
	"sync"
)

// Set is a thread-safe, ordered array of KSUID where each item exists
// exactly once.
//
// All Read/Write operations are O(1). Deletion is O(n).
type Set struct {
	items  []ID
	lookup map[ID]struct{}

	mu sync.RWMutex
}

// NewSet initializes a KSUID set for unique sets of KSUID, optionally based
// on an initial array of KSUID.
func NewSet(x ...ID) *Set {
	s := &Set{
		items:  x,
		lookup: make(map[ID]struct{}),
	}

	for _, id := range s.items {
		s.lookup[id] = struct{}{}
	}

	return s
}

// Append inserts the given ID to the end of the set, if it does not already
// exist. Returns true if item is new to the set.
//
// Complexity: O(1)
func (s *Set) Append(id ID) (appended bool) {
	s.mu.Lock()

	if _, ok := s.lookup[id]; !ok {
		s.lookup[id] = struct{}{}
		s.items = append(s.items, id)
		appended = true
	}

	s.mu.Unlock()
	return
}

// Len returns the total length of the Set.
//
// Complexity: O(1)
func (s *Set) Len() int {
	return len(s.items)
}

// Exists returns true if ID exists within the set.
//
// Complexity: O(1)
func (s *Set) Exists(id ID) (found bool) {
	s.mu.RLock()

	_, found = s.lookup[id]

	s.mu.RUnlock()
	return
}

// Delete removes ID if it exists within the set.
//
// Complexity: O(n)
func (s *Set) Delete(id ID) {
	s.mu.Lock()

	for i, x := range s.items {
		if x.Equal(id) {
			copy(s.items[i:], s.items[i+1:])
			s.items[len(s.items)-1] = ID{}
			s.items = s.items[:len(s.items)-1]
			break
		}
	}

	delete(s.lookup, id)

	s.mu.Unlock()
}

// Iterator goes over every item currently in the set in a thread-safe way.
type Iterator struct {
	s *Set

	v ID
	i int
}

// Next returns true if there is at least one more KSUID in the set
// available for iteration.
func (i *Iterator) Next() (more bool) {
	i.s.mu.RLock()

	if i.i >= len(i.s.items) {
		i.v = ID{}
		more = false
	} else {
		i.v = i.s.items[i.i]
		i.i++
		more = true
	}

	i.s.mu.RUnlock()

	return
}

// Value returns the next iterated ID.
func (i *Iterator) Value() ID {
	return i.v
}

// Iter returns a new Iterator for going over every item in
// a thread-safe manor.
func (s *Set) Iter() *Iterator {
	return &Iterator{
		s: s,
	}
}
