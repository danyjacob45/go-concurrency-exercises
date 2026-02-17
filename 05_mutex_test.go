package concurrency

import (
	"sync"
	"testing"
)

// =============================================================================
// Tests for Exercise 2.2: sync.Mutex vs sync.RWMutex
// Run with: go test -v -run Test05
// =============================================================================

// --- PART 1: Thread-Safe Counter ---

func TestCounter_Basic(t *testing.T) {
	c := NewCounter()
	if c == nil {
		t.Fatal("NewCounter returned nil")
	}

	if c.Value() != 0 {
		t.Errorf("new counter Value() = %d; want 0", c.Value())
	}

	c.Inc()
	c.Inc()
	c.Inc()
	if c.Value() != 3 {
		t.Errorf("after 3 Inc(), Value() = %d; want 3", c.Value())
	}

	c.Dec()
	if c.Value() != 2 {
		t.Errorf("after Dec(), Value() = %d; want 2", c.Value())
	}
}

func TestCounter_Concurrent(t *testing.T) {
	c := NewCounter()
	if c == nil {
		t.Fatal("NewCounter returned nil")
	}

	var wg sync.WaitGroup
	n := 1000

	// Increment from many goroutines
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.Inc()
		}()
	}
	wg.Wait()

	if c.Value() != n {
		t.Errorf("after %d concurrent Inc(), Value() = %d; want %d", n, c.Value(), n)
	}
}

// --- PART 2: Thread-Safe Cache ---

func TestCache_Basic(t *testing.T) {
	c := NewCache()
	if c == nil {
		t.Fatal("NewCache returned nil")
	}

	// Empty cache
	if c.Len() != 0 {
		t.Errorf("new cache Len() = %d; want 0", c.Len())
	}

	_, ok := c.Get("key1")
	if ok {
		t.Error("Get on empty cache returned true")
	}

	// Set and Get
	c.Set("key1", "value1")
	v, ok := c.Get("key1")
	if !ok || v != "value1" {
		t.Errorf("Get(key1) = (%q, %v); want (\"value1\", true)", v, ok)
	}

	if c.Len() != 1 {
		t.Errorf("Len() = %d; want 1", c.Len())
	}

	// Delete
	c.Delete("key1")
	_, ok = c.Get("key1")
	if ok {
		t.Error("Get after Delete returned true")
	}

	if c.Len() != 0 {
		t.Errorf("Len() after delete = %d; want 0", c.Len())
	}
}

func TestCache_Concurrent(t *testing.T) {
	c := NewCache()
	if c == nil {
		t.Fatal("NewCache returned nil")
	}

	var wg sync.WaitGroup

	// Concurrent writers
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			key := string(rune('a' + n%26))
			c.Set(key, "value")
		}(i)
	}

	// Concurrent readers
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			key := string(rune('a' + n%26))
			c.Get(key)
		}(i)
	}

	wg.Wait()
	// If no race detector errors, the test passes
}

// --- PART 3: Transfer with Lock Ordering ---

func TestTransfer_Basic(t *testing.T) {
	a1 := NewAccount(1, 100)
	a2 := NewAccount(2, 50)

	ok := Transfer(a1, a2, 30)
	if !ok {
		t.Error("Transfer returned false; want true")
	}

	if a1.Balance() != 70 {
		t.Errorf("account1 balance = %d; want 70", a1.Balance())
	}
	if a2.Balance() != 80 {
		t.Errorf("account2 balance = %d; want 80", a2.Balance())
	}
}

func TestTransfer_InsufficientFunds(t *testing.T) {
	a1 := NewAccount(1, 50)
	a2 := NewAccount(2, 100)

	ok := Transfer(a1, a2, 100)
	if ok {
		t.Error("Transfer with insufficient funds returned true; want false")
	}

	// Balances should be unchanged
	if a1.Balance() != 50 {
		t.Errorf("account1 balance = %d; want 50 (unchanged)", a1.Balance())
	}
	if a2.Balance() != 100 {
		t.Errorf("account2 balance = %d; want 100 (unchanged)", a2.Balance())
	}
}

func TestTransfer_ConcurrentNoDeadlock(t *testing.T) {
	a1 := NewAccount(1, 10000)
	a2 := NewAccount(2, 10000)

	var wg sync.WaitGroup

	// Transfer in both directions concurrently to test for deadlocks
	for i := 0; i < 100; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			Transfer(a1, a2, 1)
		}()
		go func() {
			defer wg.Done()
			Transfer(a2, a1, 1)
		}()
	}

	wg.Wait()

	// Total money should be conserved
	total := a1.Balance() + a2.Balance()
	if total != 20000 {
		t.Errorf("total balance = %d; want 20000 (money conservation violated)", total)
	}
}

// --- PART 4: SafeSlice ---

func TestSafeSlice_Basic(t *testing.T) {
	s := NewSafeSlice()
	if s == nil {
		t.Fatal("NewSafeSlice returned nil")
	}

	if s.Len() != 0 {
		t.Errorf("new SafeSlice Len() = %d; want 0", s.Len())
	}

	s.Append(10)
	s.Append(20)
	s.Append(30)

	if s.Len() != 3 {
		t.Errorf("Len() = %d; want 3", s.Len())
	}

	if v := s.Get(0); v != 10 {
		t.Errorf("Get(0) = %d; want 10", v)
	}
	if v := s.Get(2); v != 30 {
		t.Errorf("Get(2) = %d; want 30", v)
	}
}

func TestSafeSlice_Concurrent(t *testing.T) {
	s := NewSafeSlice()
	if s == nil {
		t.Fatal("NewSafeSlice returned nil")
	}

	var wg sync.WaitGroup
	n := 100

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(val int) {
			defer wg.Done()
			s.Append(val)
		}(i)
	}

	wg.Wait()

	if s.Len() != n {
		t.Errorf("after %d concurrent Appends, Len() = %d; want %d", n, s.Len(), n)
	}
}

// --- Conceptual tests ---

func TestMutexConcept_RWMutex(t *testing.T) {
	t.Log("CONCEPT: RWMutex allows multiple concurrent readers OR one exclusive writer")
	t.Log("         Use for read-heavy workloads where reads outnumber writes")
}
