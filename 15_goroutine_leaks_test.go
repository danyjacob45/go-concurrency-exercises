package concurrency

import (
	"context"
	"runtime"
	"testing"
	"time"
)

// =============================================================================
// Tests for Exercise 4.3: Goroutine Leaks
// Run with: go test -v -run Test15
// =============================================================================

// --- PART 1: Detecting Leaks ---

func TestCountGoroutines(t *testing.T) {
	count := CountGoroutines()
	if count <= 0 {
		t.Errorf("CountGoroutines() = %d; want > 0", count)
	}
}

func TestFixedFunction_NoLeak(t *testing.T) {
	before := runtime.NumGoroutine()
	FixedFunction()
	time.Sleep(50 * time.Millisecond) // Let goroutines settle
	after := runtime.NumGoroutine()

	// Should not increase goroutine count
	if after > before+1 { // Allow 1 for timing
		t.Errorf("goroutines before=%d, after=%d; possible leak", before, after)
	}
}

// --- PART 2: FetchFirstSafe ---

func TestFetchFirstSafe(t *testing.T) {
	fetchers := []func() string{
		func() string {
			time.Sleep(100 * time.Millisecond)
			return "slow1"
		},
		func() string {
			time.Sleep(10 * time.Millisecond)
			return "fast"
		},
		func() string {
			time.Sleep(200 * time.Millisecond)
			return "slow2"
		},
	}

	before := runtime.NumGoroutine()
	result := FetchFirstSafe(fetchers)
	time.Sleep(300 * time.Millisecond) // Wait for all fetchers to complete

	if result == "" {
		t.Error("FetchFirstSafe returned empty string")
	}

	after := runtime.NumGoroutine()
	if after > before+1 {
		t.Errorf("possible goroutine leak: before=%d, after=%d", before, after)
	}
}

func TestFetchFirstWithCancel(t *testing.T) {
	fetchers := []func(context.Context) string{
		func(ctx context.Context) string {
			select {
			case <-time.After(100 * time.Millisecond):
				return "slow"
			case <-ctx.Done():
				return ""
			}
		},
		func(ctx context.Context) string {
			return "fast"
		},
	}

	ctx := context.Background()
	result := FetchFirstWithCancel(ctx, fetchers)

	if result == "" {
		t.Error("FetchFirstWithCancel returned empty string")
	}
}

// --- PART 3: Worker with Done/Context ---

func TestWorkerWithDone(t *testing.T) {
	done := make(chan struct{})
	ch := WorkerWithDone(done)
	if ch == nil {
		t.Fatal("WorkerWithDone returned nil channel")
	}

	// Read a few values
	for i := 0; i < 3; i++ {
		select {
		case <-ch:
			// Got value
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for worker value")
		}
	}

	// Stop the worker
	close(done)
	time.Sleep(50 * time.Millisecond)

	// Channel should eventually be closed or stop producing
}

func TestWorkerWithContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	ch := WorkerWithContext(ctx)
	if ch == nil {
		t.Fatal("WorkerWithContext returned nil channel")
	}

	// Read a few values
	for i := 0; i < 3; i++ {
		select {
		case <-ch:
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for worker value")
		}
	}

	// Cancel and verify worker stops
	before := runtime.NumGoroutine()
	cancel()
	time.Sleep(50 * time.Millisecond)
	after := runtime.NumGoroutine()

	if after > before {
		t.Errorf("goroutines increased after cancel: before=%d, after=%d", before, after)
	}
}

// --- PART 4: GenerateWithLimit ---

func TestGenerateWithLimit(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := GenerateWithLimit(ctx, []int{1, 2, 3}, 5)
	if ch == nil {
		t.Fatal("GenerateWithLimit returned nil channel")
	}

	var values []int
	for v := range ch {
		values = append(values, v)
	}

	if len(values) != 5 {
		t.Errorf("GenerateWithLimit produced %d values; want 5", len(values))
	}
}

func TestGenerateWithLimit_Cancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	ch := GenerateWithLimit(ctx, []int{1, 2, 3}, 1000)
	if ch == nil {
		t.Fatal("GenerateWithLimit returned nil channel")
	}

	// Read a few
	<-ch
	<-ch

	// Cancel before all values produced
	cancel()

	// Should stop eventually
	count := 2
	for range ch {
		count++
	}

	if count >= 1000 {
		t.Error("GenerateWithLimit did not stop after context cancellation")
	}
}

// --- PART 5: Timer/Ticker ---

func TestSafeTimer(t *testing.T) {
	result := SafeTimer()
	if result == "" {
		t.Error("SafeTimer returned empty string")
	}
}

func TestSafeTicker(t *testing.T) {
	count := SafeTicker()
	// Should tick a few times in ~50ms with 10ms interval
	if count < 1 {
		t.Errorf("SafeTicker counted %d ticks; want >= 1", count)
	}
}

// --- PART 6: RunWithLeakCheck ---

func TestRunWithLeakCheck_NoLeak(t *testing.T) {
	ok := RunWithLeakCheck(func() {
		// No goroutine leak
		x := 1 + 1
		_ = x
	})

	if !ok {
		t.Error("RunWithLeakCheck reported leak for non-leaking function")
	}
}

// --- CHALLENGE: SafePipeline ---

func TestSafePipeline(t *testing.T) {
	ctx := context.Background()
	input := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	before := runtime.NumGoroutine()
	results := SafePipeline(ctx, input, 3)
	time.Sleep(50 * time.Millisecond)
	after := runtime.NumGoroutine()

	if results == nil {
		t.Fatal("SafePipeline returned nil")
	}

	if len(results) != 3 {
		t.Errorf("SafePipeline returned %d results; want 3", len(results))
	}

	// Each result should be doubled
	for _, v := range results {
		if v%2 != 0 {
			t.Errorf("expected doubled value, got %d", v)
		}
	}

	if after > before+1 {
		t.Errorf("possible goroutine leak: before=%d, after=%d", before, after)
	}
}

// --- Conceptual tests ---

func TestGoroutineLeakConcept(t *testing.T) {
	t.Log("CONCEPT: Goroutines are NOT garbage collected - they must exit on their own")
	t.Log("         Common causes: blocked on channel, waiting for lock, infinite loop")
	t.Log("         Prevention: context cancellation, done channels, timeouts")
	t.Log("         Detection: runtime.NumGoroutine(), pprof, goleak library")
}
