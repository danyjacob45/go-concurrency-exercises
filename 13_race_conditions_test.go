package concurrency

import (
	"sync"
	"testing"
)

// =============================================================================
// Tests for Exercise 4.1: Race Condition Detection
// Run with: go test -v -run Test13
// Race detection: go test -race -run Test13
// =============================================================================

// --- PART 1: Fix the Races ---

func TestRaceCounterFix(t *testing.T) {
	result := RaceCounterFix()
	if result != 1000 {
		t.Errorf("RaceCounterFix() = %d; want 1000", result)
	}
}

func TestRaceCounterFix_Consistent(t *testing.T) {
	// Run multiple times to ensure consistency
	for i := 0; i < 5; i++ {
		result := RaceCounterFix()
		if result != 1000 {
			t.Errorf("iteration %d: RaceCounterFix() = %d; want 1000", i, result)
		}
	}
}

func TestRaceSliceFix(t *testing.T) {
	result := RaceSliceFix()
	if result == nil {
		t.Fatal("RaceSliceFix returned nil")
	}
	if len(result) != 100 {
		t.Errorf("RaceSliceFix() has %d elements; want 100", len(result))
	}

	// Check all values 0-99 are present
	seen := make(map[int]bool)
	for _, v := range result {
		seen[v] = true
	}
	if len(seen) != 100 {
		t.Errorf("RaceSliceFix has %d unique elements; want 100", len(seen))
	}
}

func TestRaceMapFix(t *testing.T) {
	result := RaceMapFix()
	if result == nil {
		t.Fatal("RaceMapFix returned nil")
	}
	if len(result) != 100 {
		t.Errorf("RaceMapFix has %d entries; want 100", len(result))
	}

	for i := 0; i < 100; i++ {
		expected := i * i
		if result[i] != expected {
			t.Errorf("result[%d] = %d; want %d", i, result[i], expected)
		}
	}
}

// --- PART 2: SafeCounter ---

func TestSafeCounter_Basic(t *testing.T) {
	c := NewSafeCounter()
	if c == nil {
		t.Fatal("NewSafeCounter returned nil")
	}

	if c.Value() != 0 {
		t.Errorf("new SafeCounter Value() = %d; want 0", c.Value())
	}

	ok := c.IncrementIfBelow(5)
	if !ok {
		t.Error("IncrementIfBelow(5) returned false; value is 0")
	}
	if c.Value() != 1 {
		t.Errorf("Value() = %d; want 1", c.Value())
	}
}

func TestSafeCounter_Limit(t *testing.T) {
	c := NewSafeCounter()
	if c == nil {
		t.Fatal("NewSafeCounter returned nil")
	}

	// Increment up to 3
	for i := 0; i < 3; i++ {
		if !c.IncrementIfBelow(3) {
			t.Errorf("IncrementIfBelow(3) returned false at value %d", i)
		}
	}

	// Should not increment past 3
	if c.IncrementIfBelow(3) {
		t.Error("IncrementIfBelow(3) returned true at max value")
	}

	if c.Value() != 3 {
		t.Errorf("Value() = %d; want 3", c.Value())
	}
}

func TestSafeCounter_Concurrent(t *testing.T) {
	c := NewSafeCounter()
	if c == nil {
		t.Fatal("NewSafeCounter returned nil")
	}

	max := 100
	var wg sync.WaitGroup

	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.IncrementIfBelow(max)
		}()
	}

	wg.Wait()

	if c.Value() > max {
		t.Errorf("Value() = %d; should not exceed max %d", c.Value(), max)
	}
	if c.Value() != max {
		t.Errorf("Value() = %d; want exactly %d (all 200 goroutines tried, max allows 100)", c.Value(), max)
	}
}

// --- PART 3: Loop Variable Fix ---

func TestRaceLoopVarFix(t *testing.T) {
	result := RaceLoopVarFix()
	if result == nil {
		t.Fatal("RaceLoopVarFix returned nil")
	}
	if len(result) != 5 {
		t.Fatalf("RaceLoopVarFix has %d elements; want 5", len(result))
	}

	// Should contain 0,1,2,3,4 in some order
	seen := make(map[int]bool)
	for _, v := range result {
		seen[v] = true
	}

	for i := 0; i < 5; i++ {
		if !seen[i] {
			t.Errorf("value %d missing from results", i)
		}
	}
}

// --- PART 4: SafeStats ---

func TestSafeStats_Basic(t *testing.T) {
	s := NewSafeStats()
	if s == nil {
		t.Fatal("NewSafeStats returned nil")
	}

	s.Record(10)
	s.Record(20)
	s.Record(5)

	snap := s.Snapshot()
	if snap.Count != 3 {
		t.Errorf("Count = %d; want 3", snap.Count)
	}
	if snap.Sum != 35 {
		t.Errorf("Sum = %d; want 35", snap.Sum)
	}
	if snap.Max != 20 {
		t.Errorf("Max = %d; want 20", snap.Max)
	}
}

func TestSafeStats_Concurrent(t *testing.T) {
	s := NewSafeStats()
	if s == nil {
		t.Fatal("NewSafeStats returned nil")
	}

	var wg sync.WaitGroup
	n := 100

	for i := 1; i <= n; i++ {
		wg.Add(1)
		go func(val int) {
			defer wg.Done()
			s.Record(val)
		}(i)
	}

	wg.Wait()

	snap := s.Snapshot()
	if snap.Count != n {
		t.Errorf("Count = %d; want %d", snap.Count, n)
	}

	expectedSum := n * (n + 1) / 2
	if snap.Sum != expectedSum {
		t.Errorf("Sum = %d; want %d", snap.Sum, expectedSum)
	}

	if snap.Max != n {
		t.Errorf("Max = %d; want %d", snap.Max, n)
	}
}

// --- Conceptual tests ---

func TestRaceConditionConcept(t *testing.T) {
	t.Log("CONCEPT: A race condition occurs when goroutines access shared data")
	t.Log("         concurrently and at least one access is a write")
	t.Log("         Always run tests with: go test -race")
	t.Log("         Fix with: mutex, atomic, channels, or avoid sharing")
}
