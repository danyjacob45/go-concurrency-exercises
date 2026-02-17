package concurrency

import (
	"sync"
	"testing"
)

// =============================================================================
// Tests for Exercise 5.4: Go Memory Model and Happens-Before
// Run with: go test -v -run Test19
// =============================================================================

// --- PART 1: Visibility ---

func TestChannelVisibility(t *testing.T) {
	result := ChannelVisibility()
	if result != 42 {
		t.Errorf("ChannelVisibility() = %d; want 42", result)
	}
}

func TestMutexVisibility(t *testing.T) {
	result := MutexVisibility()
	if result != 42 {
		t.Errorf("MutexVisibility() = %d; want 42", result)
	}
}

// --- PART 2: Happens-Before ---

func TestChannelHappensBefore(t *testing.T) {
	a, b := ChannelHappensBefore()
	if a != 1 || b != 2 {
		t.Errorf("ChannelHappensBefore() = (%d, %d); want (1, 2)", a, b)
	}
}

func TestUnbufferedChannelOrder(t *testing.T) {
	result := UnbufferedChannelOrder()
	if result != 42 {
		t.Errorf("UnbufferedChannelOrder() = %d; want 42", result)
	}
}

func TestBufferedChannelOrder(t *testing.T) {
	result := BufferedChannelOrder()
	if result != 42 {
		t.Errorf("BufferedChannelOrder() = %d; want 42", result)
	}
}

// --- PART 3: Atomic Operations ---

func TestAtomicVisibility(t *testing.T) {
	result := AtomicVisibility()
	if result != 42 {
		t.Errorf("AtomicVisibility() = %d; want 42", result)
	}
}

func TestAtomicFix(t *testing.T) {
	result := AtomicFix()
	if result != 42 {
		t.Errorf("AtomicFix() = %d; want 42", result)
	}
}

func TestAtomicFix_Consistent(t *testing.T) {
	// Run multiple times - should always return 42
	for i := 0; i < 10; i++ {
		result := AtomicFix()
		if result != 42 {
			t.Errorf("iteration %d: AtomicFix() = %d; want 42", i, result)
		}
	}
}

// --- PART 4: sync.Once Guarantees ---

func TestGetSingletonInstance(t *testing.T) {
	// Reset for test (note: this depends on implementation)
	singletonInstance = nil
	singletonOnce = sync.Once{}

	instance := GetSingletonInstance()
	if instance == nil {
		t.Fatal("GetSingletonInstance returned nil")
	}

	if instance.data == nil {
		t.Error("singleton data is nil; should be initialized")
	}

	// Multiple calls should return same instance
	instance2 := GetSingletonInstance()
	if instance2 != instance {
		t.Error("second call returned different instance")
	}
}

func TestGetSingletonInstance_Concurrent(t *testing.T) {
	singletonInstance = nil
	singletonOnce = sync.Once{}

	var wg sync.WaitGroup
	instances := make([]*ExpensiveSingleton, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			instances[idx] = GetSingletonInstance()
		}(i)
	}

	wg.Wait()

	for i := 1; i < 10; i++ {
		if instances[i] != instances[0] {
			t.Errorf("goroutine %d got different instance", i)
		}
	}
}

// --- PART 5: Publication Safety ---

func TestSafePublish(t *testing.T) {
	SafePublish()
	cfg := GetAppConfig()
	if cfg == nil {
		t.Fatal("GetAppConfig returned nil after SafePublish")
	}

	if cfg.Timeout != 30 {
		t.Errorf("Timeout = %d; want 30", cfg.Timeout)
	}

	if len(cfg.Servers) != 3 {
		t.Errorf("Servers length = %d; want 3", len(cfg.Servers))
	}
}

// --- PART 6: LazyMap ---

func TestLazyMap_Basic(t *testing.T) {
	m := NewLazyMap()
	if m == nil {
		t.Fatal("NewLazyMap returned nil")
	}

	createCount := 0
	v := m.GetOrCreate("key1", func() any {
		createCount++
		return "value1"
	})

	if v != "value1" {
		t.Errorf("GetOrCreate = %v; want \"value1\"", v)
	}

	// Second call should not call create
	v2 := m.GetOrCreate("key1", func() any {
		createCount++
		return "value2"
	})

	if v2 != "value1" {
		t.Errorf("second GetOrCreate = %v; want \"value1\" (should reuse)", v2)
	}

	if createCount != 1 {
		t.Errorf("create called %d times; want 1", createCount)
	}
}

func TestLazyMap_Concurrent(t *testing.T) {
	m := NewLazyMap()
	if m == nil {
		t.Fatal("NewLazyMap returned nil")
	}

	var wg sync.WaitGroup
	results := make([]any, 20)

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			results[idx] = m.GetOrCreate("shared-key", func() any {
				return "created-once"
			})
		}(i)
	}

	wg.Wait()

	for i, v := range results {
		if v != "created-once" {
			t.Errorf("goroutine %d got %v; want \"created-once\"", i, v)
		}
	}
}

// --- CHALLENGE: AtomicStack ---

func TestAtomicStack_Basic(t *testing.T) {
	s := NewAtomicStack()
	if s == nil {
		t.Fatal("NewAtomicStack returned nil")
	}

	// Empty pop
	_, ok := s.Pop()
	if ok {
		t.Error("Pop on empty stack returned true")
	}

	s.Push(1)
	s.Push(2)
	s.Push(3)

	// LIFO order
	v, ok := s.Pop()
	if !ok || v != 3 {
		t.Errorf("Pop() = (%v, %v); want (3, true)", v, ok)
	}

	v, ok = s.Pop()
	if !ok || v != 2 {
		t.Errorf("Pop() = (%v, %v); want (2, true)", v, ok)
	}

	v, ok = s.Pop()
	if !ok || v != 1 {
		t.Errorf("Pop() = (%v, %v); want (1, true)", v, ok)
	}

	// Empty again
	_, ok = s.Pop()
	if ok {
		t.Error("Pop on empty stack returned true")
	}
}

func TestAtomicStack_Concurrent(t *testing.T) {
	s := NewAtomicStack()
	if s == nil {
		t.Fatal("NewAtomicStack returned nil")
	}

	var wg sync.WaitGroup
	n := 100

	// Concurrent pushes
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(val int) {
			defer wg.Done()
			s.Push(val)
		}(i)
	}

	wg.Wait()

	// Pop all
	count := 0
	for {
		_, ok := s.Pop()
		if !ok {
			break
		}
		count++
	}

	if count != n {
		t.Errorf("popped %d items; want %d", count, n)
	}
}

// --- Conceptual tests ---

func TestMemoryModelConcept(t *testing.T) {
	t.Log("CONCEPT: The Go memory model defines when writes are visible to reads")
	t.Log("         Without synchronization, there are NO visibility guarantees")
	t.Log("         Synchronization (channels, mutex, atomic) establishes happens-before")
	t.Log("         Always use the race detector: go test -race")
}
