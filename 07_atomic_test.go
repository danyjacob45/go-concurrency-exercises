package concurrency

import (
	"sync"
	"testing"
)

// =============================================================================
// Tests for Exercise 2.4: Atomic Operations
// Run with: go test -v -run Test07
// =============================================================================

// --- PART 1: AtomicCounter ---

func TestAtomicCounter_Basic(t *testing.T) {
	c := NewAtomicCounter()
	if c == nil {
		t.Fatal("NewAtomicCounter returned nil")
	}

	if c.Value() != 0 {
		t.Errorf("new counter Value() = %d; want 0", c.Value())
	}

	v := c.Inc()
	if v != 1 {
		t.Errorf("Inc() returned %d; want 1", v)
	}
	if c.Value() != 1 {
		t.Errorf("Value() = %d; want 1", c.Value())
	}

	v = c.Dec()
	if v != 0 {
		t.Errorf("Dec() returned %d; want 0", v)
	}

	v = c.Add(10)
	if v != 10 {
		t.Errorf("Add(10) returned %d; want 10", v)
	}

	v = c.Add(-3)
	if v != 7 {
		t.Errorf("Add(-3) returned %d; want 7", v)
	}
}

func TestAtomicCounter_Concurrent(t *testing.T) {
	c := NewAtomicCounter()
	if c == nil {
		t.Fatal("NewAtomicCounter returned nil")
	}

	var wg sync.WaitGroup
	n := 1000

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.Inc()
		}()
	}

	wg.Wait()

	if c.Value() != int64(n) {
		t.Errorf("after %d concurrent Inc(), Value() = %d; want %d", n, c.Value(), n)
	}
}

// --- PART 2: MaxTracker ---

func TestMaxTracker_Basic(t *testing.T) {
	m := NewMaxTracker()
	if m == nil {
		t.Fatal("NewMaxTracker returned nil")
	}

	if m.Max() != 0 {
		t.Errorf("new tracker Max() = %d; want 0", m.Max())
	}

	v := m.Update(10)
	if v != 10 {
		t.Errorf("Update(10) returned %d; want 10", v)
	}

	v = m.Update(5) // Less than current max
	if v != 10 {
		t.Errorf("Update(5) returned %d; want 10 (unchanged)", v)
	}

	v = m.Update(20)
	if v != 20 {
		t.Errorf("Update(20) returned %d; want 20", v)
	}

	if m.Max() != 20 {
		t.Errorf("Max() = %d; want 20", m.Max())
	}
}

func TestMaxTracker_Concurrent(t *testing.T) {
	m := NewMaxTracker()
	if m == nil {
		t.Fatal("NewMaxTracker returned nil")
	}

	var wg sync.WaitGroup

	for i := int64(1); i <= 1000; i++ {
		wg.Add(1)
		go func(val int64) {
			defer wg.Done()
			m.Update(val)
		}(i)
	}

	wg.Wait()

	if m.Max() != 1000 {
		t.Errorf("Max() = %d; want 1000", m.Max())
	}
}

// --- PART 3: ConfigStore ---

func TestConfigStore_Basic(t *testing.T) {
	initial := &ServerConfig{Host: "localhost", Port: 8080, Timeout: 30}
	cs := NewConfigStore(initial)
	if cs == nil {
		t.Fatal("NewConfigStore returned nil")
	}

	cfg := cs.Load()
	if cfg == nil {
		t.Fatal("Load() returned nil")
	}
	if cfg.Host != "localhost" || cfg.Port != 8080 {
		t.Errorf("Load() = {%q, %d, %d}; want {\"localhost\", 8080, 30}", cfg.Host, cfg.Port, cfg.Timeout)
	}

	// Update config
	newCfg := &ServerConfig{Host: "0.0.0.0", Port: 9090, Timeout: 60}
	cs.Store(newCfg)

	loaded := cs.Load()
	if loaded.Host != "0.0.0.0" || loaded.Port != 9090 {
		t.Errorf("after Store, Load() = {%q, %d}; want {\"0.0.0.0\", 9090}", loaded.Host, loaded.Port)
	}
}

func TestConfigStore_Concurrent(t *testing.T) {
	initial := &ServerConfig{Host: "localhost", Port: 8080, Timeout: 30}
	cs := NewConfigStore(initial)
	if cs == nil {
		t.Fatal("NewConfigStore returned nil")
	}

	var wg sync.WaitGroup

	// Concurrent readers
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cfg := cs.Load()
			if cfg == nil {
				t.Error("Load() returned nil during concurrent access")
			}
		}()
	}

	// Concurrent writer
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			cs.Store(&ServerConfig{Host: "host", Port: n, Timeout: n})
		}(i)
	}

	wg.Wait()
}

// --- PART 4: SpinLock ---

func TestSpinLock_MutualExclusion(t *testing.T) {
	var sl SpinLock
	counter := 0
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sl.Lock()
			counter++
			sl.Unlock()
		}()
	}

	wg.Wait()

	if counter != 100 {
		t.Errorf("counter = %d; want 100", counter)
	}
}

// --- Conceptual tests ---

func TestAtomicConcept(t *testing.T) {
	t.Log("CONCEPT: Atomic operations are faster than mutexes for simple operations")
	t.Log("         Use for counters, flags, and simple state")
	t.Log("         Use CAS loops for more complex lock-free algorithms")
}
