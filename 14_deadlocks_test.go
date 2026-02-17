package concurrency

import (
	"sort"
	"testing"
	"time"
)

// =============================================================================
// Tests for Exercise 4.2: Deadlock Scenarios
// Run with: go test -v -run Test14
// =============================================================================

// Helper: run function with timeout to detect deadlocks
func runWithTimeout(t *testing.T, name string, timeout time.Duration, fn func()) {
	t.Helper()
	done := make(chan struct{})
	go func() {
		fn()
		close(done)
	}()

	select {
	case <-done:
		// OK
	case <-time.After(timeout):
		t.Fatalf("%s did not complete within %v (possible deadlock)", name, timeout)
	}
}

// --- PART 1: Channel Deadlock Fixes ---

func TestFixUnbufferedSend(t *testing.T) {
	runWithTimeout(t, "FixUnbufferedSend", time.Second, func() {
		result := FixUnbufferedSend()
		if result != 42 {
			t.Errorf("FixUnbufferedSend() = %d; want 42", result)
		}
	})
}

func TestFixReceiveNoSend(t *testing.T) {
	runWithTimeout(t, "FixReceiveNoSend", time.Second, func() {
		result := FixReceiveNoSend()
		if result != 42 {
			t.Errorf("FixReceiveNoSend() = %d; want 42", result)
		}
	})
}

func TestFixBufferedFull(t *testing.T) {
	runWithTimeout(t, "FixBufferedFull", time.Second, func() {
		result := FixBufferedFull()
		if result == nil {
			t.Fatal("FixBufferedFull returned nil")
		}

		sort.Ints(result)
		expected := []int{1, 2, 3}
		if len(result) != 3 {
			t.Fatalf("FixBufferedFull returned %d values; want 3", len(result))
		}
		for i, v := range result {
			if v != expected[i] {
				t.Errorf("result[%d] = %d; want %d", i, v, expected[i])
			}
		}
	})
}

// --- PART 2: WaitGroup Deadlock Fixes ---

func TestFixWaitGroupNotDone(t *testing.T) {
	runWithTimeout(t, "FixWaitGroupNotDone", time.Second, func() {
		FixWaitGroupNotDone()
	})
}

func TestFixWaitGroupAddInGoroutine(t *testing.T) {
	runWithTimeout(t, "FixWaitGroupAddInGoroutine", time.Second, func() {
		FixWaitGroupAddInGoroutine()
	})
}

// --- PART 3: Mutex Deadlock Fixes ---

func TestFixMutexRecursive(t *testing.T) {
	runWithTimeout(t, "FixMutexRecursive", time.Second, func() {
		FixMutexRecursive()
	})
}

func TestFixMutexOrdering(t *testing.T) {
	// Run multiple times since deadlocks are timing-dependent
	for i := 0; i < 10; i++ {
		runWithTimeout(t, "FixMutexOrdering", time.Second, func() {
			FixMutexOrdering()
		})
	}
}

// --- PART 4: Channel + WaitGroup Fix ---

func TestFixChannelWaitGroup(t *testing.T) {
	runWithTimeout(t, "FixChannelWaitGroup", 2*time.Second, func() {
		result := FixChannelWaitGroup()
		if result == nil {
			t.Fatal("FixChannelWaitGroup returned nil")
		}

		if len(result) != 10 {
			t.Errorf("FixChannelWaitGroup returned %d values; want 10", len(result))
		}

		// Should contain 0-9
		sort.Ints(result)
		for i, v := range result {
			if v != i {
				t.Errorf("sorted result[%d] = %d; want %d", i, v, i)
			}
		}
	})
}

// --- PART 5: Select Fix ---

func TestFixSelectNilChannels(t *testing.T) {
	runWithTimeout(t, "FixSelectNilChannels", time.Second, func() {
		result := FixSelectNilChannels()
		if result == "" {
			t.Error("FixSelectNilChannels returned empty string")
		}
	})
}

// --- CHALLENGE: ProcessItemsFix ---

func TestProcessItemsFix_SmallInput(t *testing.T) {
	runWithTimeout(t, "ProcessItemsFix(small)", time.Second, func() {
		items := []int{1, 2, 3}
		processor := func(n int) int { return n * 2 }

		result := ProcessItemsFix(items, processor)
		if result == nil {
			t.Fatal("ProcessItemsFix returned nil")
		}

		sort.Ints(result)
		expected := []int{2, 4, 6}
		if len(result) != len(expected) {
			t.Fatalf("got %d results; want %d", len(result), len(expected))
		}
		for i, v := range result {
			if v != expected[i] {
				t.Errorf("sorted result[%d] = %d; want %d", i, v, expected[i])
			}
		}
	})
}

func TestProcessItemsFix_LargeInput(t *testing.T) {
	// This is the key test: >5 items should still work (original deadlocks with buffer=5)
	runWithTimeout(t, "ProcessItemsFix(large)", 2*time.Second, func() {
		items := make([]int, 20)
		for i := range items {
			items[i] = i
		}
		processor := func(n int) int { return n + 1 }

		result := ProcessItemsFix(items, processor)
		if result == nil {
			t.Fatal("ProcessItemsFix returned nil")
		}

		if len(result) != 20 {
			t.Errorf("got %d results; want 20", len(result))
		}
	})
}

func TestProcessItemsFix_Empty(t *testing.T) {
	runWithTimeout(t, "ProcessItemsFix(empty)", time.Second, func() {
		result := ProcessItemsFix([]int{}, func(n int) int { return n })
		if len(result) != 0 {
			t.Errorf("got %d results for empty input; want 0", len(result))
		}
	})
}

// --- Conceptual tests ---

func TestDeadlockConcept(t *testing.T) {
	t.Log("CONCEPT: Deadlocks occur when goroutines wait for each other forever")
	t.Log("         Go's runtime detects deadlock when ALL goroutines are blocked")
	t.Log("         Common causes: missing receiver/sender, wrong WaitGroup usage,")
	t.Log("         recursive mutex locking, AB-BA lock ordering")
}
