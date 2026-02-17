package concurrency

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// =============================================================================
// Tests for Exercise 2.3: sync.Once and sync.Cond
// Run with: go test -v -run Test06
// =============================================================================

// --- PART 1: sync.Once - Lazy Initialization ---

func TestResourceManager_LazyInit(t *testing.T) {
	callCount := int32(0)
	createFn := func() *ExpensiveResource {
		atomic.AddInt32(&callCount, 1)
		return &ExpensiveResource{data: "initialized"}
	}

	rm := NewResourceManager(createFn)
	if rm == nil {
		t.Fatal("NewResourceManager returned nil")
	}

	// First Get should trigger creation
	r := rm.Get()
	if r == nil {
		t.Fatal("Get() returned nil")
	}
	if r.data != "initialized" {
		t.Errorf("resource data = %q; want \"initialized\"", r.data)
	}

	// Second Get should return same instance
	r2 := rm.Get()
	if r2 != r {
		t.Error("second Get() returned different instance; want same pointer")
	}

	// createFn should only be called once
	if atomic.LoadInt32(&callCount) != 1 {
		t.Errorf("createFn called %d times; want 1", callCount)
	}
}

func TestResourceManager_ConcurrentInit(t *testing.T) {
	callCount := int32(0)
	createFn := func() *ExpensiveResource {
		atomic.AddInt32(&callCount, 1)
		time.Sleep(10 * time.Millisecond) // Simulate expensive creation
		return &ExpensiveResource{data: "shared"}
	}

	rm := NewResourceManager(createFn)
	if rm == nil {
		t.Fatal("NewResourceManager returned nil")
	}

	var wg sync.WaitGroup
	results := make([]*ExpensiveResource, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			results[idx] = rm.Get()
		}(i)
	}

	wg.Wait()

	// All should get the same instance
	for i := 1; i < 10; i++ {
		if results[i] != results[0] {
			t.Errorf("goroutine %d got different instance", i)
		}
	}

	if atomic.LoadInt32(&callCount) != 1 {
		t.Errorf("createFn called %d times; want 1", callCount)
	}
}

// --- PART 2: ConfigLoader ---

func TestConfigLoader(t *testing.T) {
	callCount := int32(0)
	loader := func() *Config {
		atomic.AddInt32(&callCount, 1)
		return &Config{
			DatabaseURL: "postgres://localhost",
			APIKey:      "secret",
			Debug:       true,
		}
	}

	cl := NewConfigLoader(loader)
	if cl == nil {
		t.Fatal("NewConfigLoader returned nil")
	}

	// Before loading
	if cl.Get() != nil {
		t.Error("Get() before Load() should return nil")
	}

	// Load
	cl.Load()
	cfg := cl.Get()
	if cfg == nil {
		t.Fatal("Get() after Load() returned nil")
	}
	if cfg.DatabaseURL != "postgres://localhost" {
		t.Errorf("DatabaseURL = %q; want \"postgres://localhost\"", cfg.DatabaseURL)
	}

	// Loading again should not call loader
	cl.Load()
	if atomic.LoadInt32(&callCount) != 1 {
		t.Errorf("loader called %d times; want 1", callCount)
	}
}

// --- PART 3: BoundedQueue ---

func TestBoundedQueue_Basic(t *testing.T) {
	q := NewBoundedQueue(3)
	if q == nil {
		t.Fatal("NewBoundedQueue returned nil")
	}

	if q.Len() != 0 {
		t.Errorf("new queue Len() = %d; want 0", q.Len())
	}

	q.Put(1)
	q.Put(2)
	q.Put(3)

	if q.Len() != 3 {
		t.Errorf("after 3 Puts, Len() = %d; want 3", q.Len())
	}

	v := q.Get()
	if v != 1 {
		t.Errorf("first Get() = %d; want 1", v)
	}

	v = q.Get()
	if v != 2 {
		t.Errorf("second Get() = %d; want 2", v)
	}
}

func TestBoundedQueue_BlocksWhenFull(t *testing.T) {
	q := NewBoundedQueue(2)
	if q == nil {
		t.Fatal("NewBoundedQueue returned nil")
	}

	q.Put(1)
	q.Put(2)

	// Third Put should block until a Get happens
	done := make(chan bool)
	go func() {
		q.Put(3) // Should block
		done <- true
	}()

	select {
	case <-done:
		t.Error("Put did not block on full queue")
	case <-time.After(50 * time.Millisecond):
		// Expected: Put is blocking
	}

	// Now consume one
	v := q.Get()
	if v != 1 {
		t.Errorf("Get() = %d; want 1", v)
	}

	// Put should now complete
	select {
	case <-done:
		// Good
	case <-time.After(100 * time.Millisecond):
		t.Error("Put still blocked after Get freed space")
	}
}

func TestBoundedQueue_BlocksWhenEmpty(t *testing.T) {
	q := NewBoundedQueue(5)
	if q == nil {
		t.Fatal("NewBoundedQueue returned nil")
	}

	done := make(chan int)
	go func() {
		done <- q.Get() // Should block
	}()

	select {
	case <-done:
		t.Error("Get did not block on empty queue")
	case <-time.After(50 * time.Millisecond):
		// Expected: Get is blocking
	}

	q.Put(42)

	select {
	case v := <-done:
		if v != 42 {
			t.Errorf("Get() = %d; want 42", v)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Get still blocked after Put added item")
	}
}

func TestBoundedQueue_ProducerConsumer(t *testing.T) {
	q := NewBoundedQueue(5)
	if q == nil {
		t.Fatal("NewBoundedQueue returned nil")
	}

	count := 50
	var wg sync.WaitGroup

	// Producer
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < count; i++ {
			q.Put(i)
		}
	}()

	// Consumer
	results := make([]int, 0, count)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < count; i++ {
			results = append(results, q.Get())
		}
	}()

	wg.Wait()

	if len(results) != count {
		t.Errorf("received %d items; want %d", len(results), count)
	}
}

// --- PART 4: Barrier ---

func TestBarrier(t *testing.T) {
	n := 5
	b := NewBarrier(n)
	if b == nil {
		t.Fatal("NewBarrier returned nil")
	}

	arrivals := make([]int, n)
	var wg sync.WaitGroup

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			arrivals[idx] = b.Wait()
		}(i)
	}

	wg.Wait()

	// Each goroutine should get a unique arrival order 1..n
	seen := make(map[int]bool)
	for _, v := range arrivals {
		if v < 1 || v > n {
			t.Errorf("arrival order %d out of range [1, %d]", v, n)
		}
		seen[v] = true
	}

	if len(seen) != n {
		t.Errorf("got %d unique arrival orders; want %d", len(seen), n)
	}
}

// --- Conceptual tests ---

func TestOnceConcept(t *testing.T) {
	t.Log("CONCEPT: sync.Once ensures a function runs exactly once")
	t.Log("         Even if Do() is called concurrently from many goroutines")
	t.Log("         Subsequent calls return immediately without running f")
}

func TestCondConcept(t *testing.T) {
	t.Log("CONCEPT: sync.Cond provides condition-based waking")
	t.Log("         Signal() wakes ONE waiter, Broadcast() wakes ALL")
	t.Log("         Always check condition in a loop after Wait() returns")
}
