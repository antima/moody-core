package mqtt

import "sync"

type Iterable interface {
	Next() (interface{}, bool)
}

type SingleIterator struct {
	nextChan chan interface{}
}

type ConcurrentSet struct {
	mutex sync.RWMutex
	set   map[interface{}]bool
}

func NewConcurrentSet() *ConcurrentSet {
	return &ConcurrentSet{
		set: make(map[interface{}]bool),
	}
}

func (concurrentSet *ConcurrentSet) Add(elem interface{}) {
	concurrentSet.mutex.Lock()
	defer concurrentSet.mutex.Unlock()
	concurrentSet.set[elem] = true
}

func (concurrentSet *ConcurrentSet) Remove(elem interface{}) {
	concurrentSet.mutex.Lock()
	defer concurrentSet.mutex.Unlock()
	delete(concurrentSet.set, elem)
}

func (concurrentSet *ConcurrentSet) Contains(elem interface{}) bool {
	concurrentSet.mutex.RLock()
	defer concurrentSet.mutex.RUnlock()
	_, contains := concurrentSet.set[elem]
	return contains
}

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

func (concurrentSet *ConcurrentSet) Size() int {
	return len(concurrentSet.set)
}

func (concurrentSet *ConcurrentSet) ToSlice() []interface{} {
	concurrentSet.mutex.RLock()
	defer concurrentSet.mutex.RUnlock()
	var services []interface{}
	for elem := range concurrentSet.set {
		services = append(services, elem)
	}
	return services
}

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

func (iterator *SingleIterator) Next() (interface{}, bool) {
	nextElement, open := <-iterator.nextChan
	return nextElement, !open
}
