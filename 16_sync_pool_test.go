package concurrency

import (
	"bytes"
	"sync"
	"testing"
)

// =============================================================================
// Tests for Exercise 5.1: sync.Pool for Object Reuse
// Run with: go test -v -run Test16
// =============================================================================

// --- PART 1: BufferPool ---

func TestBufferPool_Basic(t *testing.T) {
	bp := NewBufferPool()
	if bp == nil {
		t.Fatal("NewBufferPool returned nil")
	}

	// Get a buffer
	buf := bp.Get()
	if buf == nil {
		t.Fatal("Get() returned nil")
	}

	// Write to buffer
	buf.WriteString("hello")
	if buf.String() != "hello" {
		t.Errorf("buffer content = %q; want \"hello\"", buf.String())
	}

	// Put back and get again - should be reset
	bp.Put(buf)
	buf2 := bp.Get()
	if buf2 == nil {
		t.Fatal("second Get() returned nil")
	}

	// Buffer should be reset (empty)
	if buf2.Len() != 0 {
		t.Errorf("reused buffer Len() = %d; want 0 (should be reset)", buf2.Len())
	}
}

func TestBufferPool_Concurrent(t *testing.T) {
	bp := NewBufferPool()
	if bp == nil {
		t.Fatal("NewBufferPool returned nil")
	}

	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			buf := bp.Get()
			buf.WriteString("test")
			bp.Put(buf)
		}()
	}

	wg.Wait()
}

// --- PART 2: SizedBufferPool ---

func TestSizedBufferPool_Basic(t *testing.T) {
	sp := NewSizedBufferPool(1024)
	if sp == nil {
		t.Fatal("NewSizedBufferPool returned nil")
	}

	buf := sp.Get()
	if buf == nil {
		t.Fatal("Get() returned nil")
	}

	if cap(buf) < 1024 {
		t.Errorf("buffer cap = %d; want >= 1024", cap(buf))
	}

	sp.Put(buf)
}

func TestSizedBufferPool_Concurrent(t *testing.T) {
	sp := NewSizedBufferPool(512)
	if sp == nil {
		t.Fatal("NewSizedBufferPool returned nil")
	}

	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			buf := sp.Get()
			_ = buf
			sp.Put(buf)
		}()
	}

	wg.Wait()
}

// --- PART 3: TrackedPool ---

func TestTrackedPool_Stats(t *testing.T) {
	tp := NewTrackedPool(func() any {
		return new(bytes.Buffer)
	})
	if tp == nil {
		t.Fatal("NewTrackedPool returned nil")
	}

	// First Get - should be a miss
	obj := tp.Get()
	if obj == nil {
		t.Fatal("Get() returned nil")
	}

	stats := tp.Stats()
	if stats.Gets != 1 {
		t.Errorf("Gets = %d; want 1", stats.Gets)
	}
	if stats.Misses != 1 {
		t.Errorf("Misses = %d; want 1", stats.Misses)
	}

	// Put and Get again - should be a hit
	tp.Put(obj)
	obj2 := tp.Get()
	_ = obj2

	stats = tp.Stats()
	if stats.Gets != 2 {
		t.Errorf("Gets = %d; want 2", stats.Gets)
	}
	if stats.Puts != 1 {
		t.Errorf("Puts = %d; want 1", stats.Puts)
	}
	// Note: Hits may or may not be 1 depending on GC
}

// --- PART 4: IntSlicePool ---

func TestIntSlicePool_Basic(t *testing.T) {
	p := NewIntSlicePool(100)
	if p == nil {
		t.Fatal("NewIntSlicePool returned nil")
	}

	s := p.Get()
	if s == nil {
		t.Fatal("Get() returned nil")
	}

	if len(s) != 0 {
		t.Errorf("new slice len = %d; want 0", len(s))
	}

	if cap(s) < 100 {
		t.Errorf("new slice cap = %d; want >= 100", cap(s))
	}

	// Use the slice
	s = append(s, 1, 2, 3)

	// Put back - should be reusable
	p.Put(s)

	s2 := p.Get()
	if len(s2) != 0 {
		t.Errorf("reused slice len = %d; want 0", len(s2))
	}
}

// --- PART 5: GoodConnectionPool ---

func TestGoodConnectionPool_Basic(t *testing.T) {
	factory := func(id int) *PooledConn {
		return &PooledConn{ID: id}
	}

	pool := NewGoodConnectionPool(3, factory)
	if pool == nil {
		t.Fatal("NewGoodConnectionPool returned nil")
	}
	defer pool.Close()

	// Get all connections
	c1 := pool.Get()
	c2 := pool.Get()
	c3 := pool.Get()

	if c1 == nil || c2 == nil || c3 == nil {
		t.Fatal("Get() returned nil")
	}

	// Return and get again
	pool.Put(c1)
	c4 := pool.Get()
	if c4 == nil {
		t.Fatal("Get() after Put returned nil")
	}

	pool.Put(c2)
	pool.Put(c3)
	pool.Put(c4)
}

func TestGoodConnectionPool_Close(t *testing.T) {
	factory := func(id int) *PooledConn {
		return &PooledConn{ID: id}
	}

	pool := NewGoodConnectionPool(2, factory)
	if pool == nil {
		t.Fatal("NewGoodConnectionPool returned nil")
	}

	pool.Close()
	// Closing should not panic
}

// --- Conceptual tests ---

func TestSyncPoolConcept(t *testing.T) {
	t.Log("CONCEPT: sync.Pool caches allocated objects to reduce GC pressure")
	t.Log("         Objects may be removed at any time (between GC cycles)")
	t.Log("         Always reset objects before Put() to avoid stale data")
	t.Log("         NOT for connection pools - use channels instead")
}
