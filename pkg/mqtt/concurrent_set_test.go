package mqtt

import (
	"fmt"
	"reflect"
	"sync"
	"testing"
)

func TestAddContains(t *testing.T) {
	var waitGroup sync.WaitGroup
	set := NewConcurrentSet()
	waitGroup.Add(100)
	for idx := 0; idx < 100; idx++ {
		s := fmt.Sprintf("test%d", idx)
		go func(set *ConcurrentSet, str string, wg *sync.WaitGroup) {
			set.Add(s)
			wg.Done()
		}(set, s, &waitGroup)
	}
	waitGroup.Wait()

	for idx := 0; idx < 100; idx++ {
		s := fmt.Sprintf("test%d", idx)
		if !set.Contains(s) {
			t.Errorf("expected set[%s] = true, got false", s)
		}
	}
}

func TestSize(t *testing.T) {
	var waitGroup sync.WaitGroup
	set := NewConcurrentSet()
	waitGroup.Add(100)
	for idx := 0; idx < 100; idx++ {
		s := fmt.Sprintf("test%d", idx)
		go func(set *ConcurrentSet, str string, wg *sync.WaitGroup) {
			set.Add(s)
			wg.Done()
		}(set, s, &waitGroup)
	}
	waitGroup.Wait()

	if set.Size() != 100 {
		t.Errorf("expected len(set) = 100, got %d", set.Size())
	}
}

func TestAddRemove(t *testing.T) {
	var waitGroup sync.WaitGroup
	set := NewConcurrentSet()
	waitGroup.Add(100)
	for idx := 0; idx < 100; idx++ {
		s := fmt.Sprintf("test%d", idx)
		go func(set *ConcurrentSet, str string, wg *sync.WaitGroup) {
			set.Add(s)
			wg.Done()
		}(set, s, &waitGroup)
	}
	waitGroup.Wait()

	waitGroup.Add(100)
	for idx := 0; idx < 100; idx++ {
		s := fmt.Sprintf("test%d", idx)
		go func(set *ConcurrentSet, str string, wg *sync.WaitGroup) {
			set.Remove(s)
			wg.Done()
		}(set, s, &waitGroup)
	}
	waitGroup.Wait()

	if set.Size() != 0 {
		t.Errorf("expected len(set) = 0, got %d", set.Size())
	}
}

func TestDifference(t *testing.T) {
	var waitGroup sync.WaitGroup
	set1 := NewConcurrentSet()
	waitGroup.Add(100)
	for idx := 0; idx < 100; idx++ {
		s := fmt.Sprintf("test%d", idx)
		go func(set *ConcurrentSet, str string, wg *sync.WaitGroup) {
			set.Add(s)
			wg.Done()
		}(set1, s, &waitGroup)
	}
	waitGroup.Wait()

	set2 := NewConcurrentSet()
	waitGroup.Add(50)
	for idx := 0; idx < 50; idx++ {
		s := fmt.Sprintf("test%d", idx)
		go func(set *ConcurrentSet, str string, wg *sync.WaitGroup) {
			set.Add(s)
			wg.Done()
		}(set2, s, &waitGroup)
	}
	waitGroup.Wait()

	diffSet := set1.Difference(set2)
	for idx := 0; idx < diffSet.Size(); idx++ {
		s := fmt.Sprintf("test%d", idx+50)
		if !diffSet.Contains(s) {
			t.Errorf("expected Contains(%s) = true, got false", s)
		}
	}
}

func TestToSlice(t *testing.T) {
	var waitGroup sync.WaitGroup
	set := NewConcurrentSet()
	waitGroup.Add(100)
	for idx := 0; idx < 100; idx++ {
		go func(set *ConcurrentSet, idx int, wg *sync.WaitGroup) {
			set.Add(idx)
			wg.Done()
		}(set, idx, &waitGroup)
	}
	waitGroup.Wait()

	slice := set.ToSlice()
	for _, elem := range slice {
		intelem, isInt := elem.(int)
		if !isInt {
			t.Errorf("expected int type, got %v", reflect.TypeOf(elem))
		}
		if !set.Contains(intelem) {
			t.Errorf("expected Contains(%d) = true, got false", intelem)
		}
	}
}
