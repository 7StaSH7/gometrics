package pool

import (
	"sync"
	"testing"
)

// TestObject is a simple struct that implements Resetter.
type TestObject struct {
	ID          int
	Value       string
	ResetCalled bool
}

// Reset implements the Resetter interface.
func (o *TestObject) Reset() {
	o.ID = 0
	o.Value = ""
	o.ResetCalled = true
}

// TestPool_NewAndGet verifies that New creates a pool and Get retrieves objects created by the generator.
func TestPool_NewAndGet(t *testing.T) {
	counter := 0
	factory := func() *TestObject {
		counter++
		return &TestObject{ID: counter, Value: "initial"}
	}

	p := New(factory)

	obj1 := p.Get()
	if obj1 == nil {
		t.Fatal("Expected object, got nil")
	}
	if obj1.Value != "initial" {
		t.Errorf("Expected 'initial', got '%s'", obj1.Value)
	}

	obj2 := p.Get()
	if obj2 == nil {
		t.Fatal("Expected object, got nil")
	}
	if obj1.ID == obj2.ID {
		t.Error("Expected different objects for subsequent Gets on empty pool")
	}
}

// TestPool_PutResets verifies that Put calls Reset() on the object.
func TestPool_PutResets(t *testing.T) {
	p := New(func() *TestObject {
		return &TestObject{}
	})

	obj := p.Get()
	obj.ID = 42
	obj.Value = "modified"
	obj.ResetCalled = false

	p.Put(obj)

	if !obj.ResetCalled {
		t.Error("Put() did not call Reset() on the object")
	}
	if obj.ID != 0 || obj.Value != "" {
		t.Errorf("Object state was not reset. ID=%d, Value='%s'", obj.ID, obj.Value)
	}
}

// TestPool_Reuse verifies that we can retrieve an object that was put back.
func TestPool_Reuse(t *testing.T) {
	p := New(func() *TestObject {
		return &TestObject{ID: -1} // -1 marks new objects
	})

	obj := p.Get()
	obj.ID = 100
	p.Put(obj)

	gotCached := false
	for range 10 {
		obj2 := p.Get()
		if obj2.ID == 0 {
			gotCached = true
			break
		}
	}

	if !gotCached {
		t.Log("Did not reuse object (this is valid for sync.Pool behavior)")
	}
}

// TestPool_Concurrency checks for race conditions.
func TestPool_Concurrency(t *testing.T) {
	p := New(func() *TestObject {
		return &TestObject{}
	})

	const workers = 10
	const iterations = 1000

	var wg sync.WaitGroup
	wg.Add(workers)

	for range workers {
		go func() {
			defer wg.Done()
			for j := range iterations {
				obj := p.Get()

				// Simulate usage
				obj.ID = j
				obj.Value = "working"

				p.Put(obj)
			}
		}()
	}

	wg.Wait()
}

// TestPool_SliceReset test with a struct containing a slice to ensure complex resets work if implemented.
type SliceObject struct {
	Data []int
}

func (s *SliceObject) Reset() {
	s.Data = s.Data[:0]
}

func TestPool_SliceReset(t *testing.T) {
	p := New(func() *SliceObject {
		return &SliceObject{Data: make([]int, 0, 10)}
	})

	obj := p.Get()
	obj.Data = append(obj.Data, 1, 2, 3)

	capBefore := cap(obj.Data)
	p.Put(obj)

	if len(obj.Data) != 0 {
		t.Errorf("Slice was not truncated. Len: %d", len(obj.Data))
	}
	if cap(obj.Data) != capBefore {
		t.Errorf("Capacity changed. Before: %d, After: %d", capBefore, cap(obj.Data))
	}
}
