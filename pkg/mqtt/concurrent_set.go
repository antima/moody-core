package mqtt

import "sync"

// Iterable defines a type that can be iterated upon
type Iterable interface {
	Next() (interface{}, bool)
}

// SingleIterator is a type that can be iterated only in the forward
// direction
type SingleIterator struct {
	nextChan chan interface{}
}

// ConcurrentSet implements a map-based, synchronized set type
type ConcurrentSet struct {
	mutex sync.RWMutex
	set   map[interface{}]bool
}

// NewConcurrentSet creates a new set and returns it as a pointer
func NewConcurrentSet() *ConcurrentSet {
	return &ConcurrentSet{
		set: make(map[interface{}]bool),
	}
}

// Add an element to the set, this changes nothing if the element
// already exists
func (concurrentSet *ConcurrentSet) Add(elem interface{}) {
	concurrentSet.mutex.Lock()
	defer concurrentSet.mutex.Unlock()
	concurrentSet.set[elem] = true
}

// Remove an element from the set, this changes nothing if the element
// does not exist
func (concurrentSet *ConcurrentSet) Remove(elem interface{}) {
	concurrentSet.mutex.Lock()
	defer concurrentSet.mutex.Unlock()
	delete(concurrentSet.set, elem)
}

// Contains returns true if the element is present in the set
func (concurrentSet *ConcurrentSet) Contains(elem interface{}) bool {
	concurrentSet.mutex.RLock()
	defer concurrentSet.mutex.RUnlock()
	_, contains := concurrentSet.set[elem]
	return contains
}

// Difference returns a set containing elements that are contained in
// the current set but are not in the passed one
func (concurrentSet *ConcurrentSet) Difference(set *ConcurrentSet) *ConcurrentSet {
	concurrentSet.mutex.RLock()
	defer concurrentSet.mutex.RUnlock()
	diffSet := NewConcurrentSet()
	for elem := range concurrentSet.set {
		if !set.Contains(elem) {
			diffSet.Add(elem)
		}
	}
	return diffSet
}

// Size returns the number of elements in the set
func (concurrentSet *ConcurrentSet) Size() int {
	return len(concurrentSet.set)
}

// ToSlice returns a representation of the set as a slice of all
// its elements
func (concurrentSet *ConcurrentSet) ToSlice() []interface{} {
	concurrentSet.mutex.RLock()
	defer concurrentSet.mutex.RUnlock()
	var services []interface{}
	for elem := range concurrentSet.set {
		services = append(services, elem)
	}
	return services
}

// Iterator returns an iterable object that can be used to travel
// the set
func (concurrentSet *ConcurrentSet) Iterator() Iterable {
	iterator := &SingleIterator{
		nextChan: make(chan interface{}),
	}
	go func() {
		for key := range concurrentSet.set {
			iterator.nextChan <- key
		}
		close(iterator.nextChan)
	}()
	return iterator
}

// Next returns the next element in the collection that the iterator
// is travelling through. If the second returned element is false,
// the end of the collection was reached.
func (iterator *SingleIterator) Next() (interface{}, bool) {
	nextElement, open := <-iterator.nextChan
	return nextElement, !open
}
