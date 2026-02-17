package concurrency

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// =============================================================================
// Tests for Exercise 3.5: Semaphore Pattern
// Run with: go test -v -run Test12
// =============================================================================

// --- PART 1: Channel-Based Semaphore ---

func TestSemaphore_Basic(t *testing.T) {
	sem := NewSemaphore(3)
	if sem == nil {
		t.Fatal("NewSemaphore returned nil")
	}

	sem.Acquire()
	sem.Acquire()
	sem.Acquire()

	// All 3 slots taken; TryAcquire should fail
	if sem.TryAcquire() {
		t.Error("TryAcquire succeeded when semaphore should be full")
	}

	sem.Release()

	// Now TryAcquire should succeed
	if !sem.TryAcquire() {
		t.Error("TryAcquire failed after Release")
	}

	sem.Release()
	sem.Release()
	sem.Release()
}

func TestSemaphore_ConcurrencyLimit(t *testing.T) {
	sem := NewSemaphore(3)
	if sem == nil {
		t.Fatal("NewSemaphore returned nil")
	}

	var concurrent int64
	var maxConcurrent int64
	var wg sync.WaitGroup

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem.Acquire()
			defer sem.Release()

			cur := atomic.AddInt64(&concurrent, 1)
			// Track max concurrent
			for {
				old := atomic.LoadInt64(&maxConcurrent)
				if cur <= old || atomic.CompareAndSwapInt64(&maxConcurrent, old, cur) {
					break
				}
			}
			time.Sleep(10 * time.Millisecond)
			atomic.AddInt64(&concurrent, -1)
		}()
	}

	wg.Wait()

	if maxConcurrent > 3 {
		t.Errorf("max concurrent = %d; want <= 3", maxConcurrent)
	}
}

// --- PART 2: TimeoutSemaphore ---

func TestTimeoutSemaphore_AcquireTimeout_Success(t *testing.T) {
	sem := NewTimeoutSemaphore(1)
	if sem == nil {
		t.Fatal("NewTimeoutSemaphore returned nil")
	}

	ok := sem.AcquireTimeout(100 * time.Millisecond)
	if !ok {
		t.Error("AcquireTimeout failed; semaphore should have free slot")
	}
	sem.Release()
}

func TestTimeoutSemaphore_AcquireTimeout_Timeout(t *testing.T) {
	sem := NewTimeoutSemaphore(1)
	if sem == nil {
		t.Fatal("NewTimeoutSemaphore returned nil")
	}

	sem.Acquire() // Take the only slot

	start := time.Now()
	ok := sem.AcquireTimeout(50 * time.Millisecond)
	elapsed := time.Since(start)

	if ok {
		t.Error("AcquireTimeout should have timed out")
	}
	if elapsed < 40*time.Millisecond {
		t.Errorf("AcquireTimeout returned too fast (%v); expected ~50ms", elapsed)
	}
	sem.Release()
}

func TestTimeoutSemaphore_AcquireContext(t *testing.T) {
	sem := NewTimeoutSemaphore(1)
	if sem == nil {
		t.Fatal("NewTimeoutSemaphore returned nil")
	}

	sem.Acquire() // Take the only slot

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := sem.AcquireContext(ctx)
	if err == nil {
		t.Error("AcquireContext with cancelled context should return error")
	}
	sem.Release()
}

// --- PART 3: BoundedParallel ---

func TestBoundedParallel(t *testing.T) {
	fns := make([]func() int, 10)
	for i := range fns {
		val := i
		fns[i] = func() int { return val * val }
	}

	results := BoundedParallel(fns, 3)
	if results == nil {
		t.Fatal("BoundedParallel returned nil")
	}

	if len(results) != 10 {
		t.Fatalf("BoundedParallel returned %d results; want 10", len(results))
	}

	// Results should be in order
	for i, v := range results {
		expected := i * i
		if v != expected {
			t.Errorf("results[%d] = %d; want %d", i, v, expected)
		}
	}
}

func TestBoundedParallelError(t *testing.T) {
	fns := []func(context.Context) error{
		func(ctx context.Context) error { return nil },
		func(ctx context.Context) error { return context.DeadlineExceeded },
		func(ctx context.Context) error { return nil },
	}

	err := BoundedParallelError(context.Background(), fns, 2)
	if err == nil {
		t.Error("BoundedParallelError should have returned an error")
	}
}

// --- PART 4: WeightedSemaphore ---

func TestWeightedSemaphore_Basic(t *testing.T) {
	ws := NewWeightedSemaphore(10)
	if ws == nil {
		t.Fatal("NewWeightedSemaphore returned nil")
	}

	if !ws.TryAcquire(5) {
		t.Error("TryAcquire(5) failed; capacity should be 10")
	}

	if !ws.TryAcquire(5) {
		t.Error("TryAcquire(5) failed; should have 5 remaining")
	}

	if ws.TryAcquire(1) {
		t.Error("TryAcquire(1) succeeded; no capacity remaining")
	}

	ws.Release(3)

	if !ws.TryAcquire(3) {
		t.Error("TryAcquire(3) failed after Release(3)")
	}

	ws.Release(10)
}

// --- PART 5: ConnectionPool ---

func TestConnectionPool_Basic(t *testing.T) {
	factory := func(id int) *Connection { return &Connection{ID: id} }

	pool := NewConnectionPool(3, factory)
	if pool == nil {
		t.Fatal("NewConnectionPool returned nil")
	}
	defer pool.Close()

	c1 := pool.Get()
	if c1 == nil {
		t.Fatal("Get() returned nil")
	}

	c2 := pool.Get()
	if c2 == nil {
		t.Fatal("Get() returned nil")
	}

	// Return one
	pool.Put(c1)

	c3 := pool.Get()
	if c3 == nil {
		t.Fatal("Get() returned nil after Put")
	}

	pool.Put(c2)
	pool.Put(c3)
}

func TestConnectionPool_GetTimeout(t *testing.T) {
	factory := func(id int) *Connection { return &Connection{ID: id} }

	pool := NewConnectionPool(1, factory)
	if pool == nil {
		t.Fatal("NewConnectionPool returned nil")
	}
	defer pool.Close()

	c := pool.Get()

	// Pool is exhausted, GetTimeout should fail
	_, ok := pool.GetTimeout(50 * time.Millisecond)
	if ok {
		t.Error("GetTimeout succeeded on exhausted pool")
	}

	pool.Put(c)
}

// --- Conceptual tests ---

func TestSemaphoreConcept(t *testing.T) {
	t.Log("CONCEPT: Semaphore limits concurrent access to N instead of 1 (mutex)")
	t.Log("         Use buffered channel of struct{} for simple implementation")
	t.Log("         Acquire = send to channel, Release = receive from channel")
}
