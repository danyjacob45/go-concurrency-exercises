package concurrency

import (
	"context"
	"testing"
	"time"
)

// =============================================================================
// Tests for Exercise 3.3: Context and Cancellation
// Run with: go test -v -run Test10
// =============================================================================

// --- PART 1: Basic Cancellation ---

func TestDoWorkWithCancel(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	result := DoWorkWithCancel(ctx)

	// Should have done some work before cancellation
	if result <= 0 {
		t.Errorf("DoWorkWithCancel did %d units of work; want > 0", result)
	}

	// Should not run forever (timeout should have cancelled it)
	t.Logf("Completed %d units of work before cancellation", result)
}

func TestDoWorkWithCancel_ImmediateCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	result := DoWorkWithCancel(ctx)

	if result > 1 {
		t.Errorf("DoWorkWithCancel after immediate cancel did %d work; want 0 or 1", result)
	}
}

func TestSearchWithCancel_Found(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	data := []string{"apple", "banana", "cherry", "date"}
	result, found := SearchWithCancel(ctx, data, "cherry")

	if !found {
		t.Error("SearchWithCancel did not find \"cherry\"")
	}
	if result != "cherry" {
		t.Errorf("SearchWithCancel returned %q; want \"cherry\"", result)
	}
}

func TestSearchWithCancel_Cancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel before search

	data := []string{"apple", "banana", "cherry"}
	_, found := SearchWithCancel(ctx, data, "cherry")

	if found {
		t.Error("SearchWithCancel returned found=true after cancellation")
	}
}

// --- PART 2: Timeouts ---

func TestFetchWithTimeout_Success(t *testing.T) {
	fetcher := func() string {
		time.Sleep(10 * time.Millisecond)
		return "data"
	}

	result, err := FetchWithTimeout(fetcher, 1*time.Second)

	if err != nil {
		t.Errorf("FetchWithTimeout returned error: %v", err)
	}
	if result != "data" {
		t.Errorf("FetchWithTimeout = %q; want \"data\"", result)
	}
}

func TestFetchWithTimeout_Timeout(t *testing.T) {
	fetcher := func() string {
		time.Sleep(1 * time.Second)
		return "slow data"
	}

	_, err := FetchWithTimeout(fetcher, 50*time.Millisecond)

	if err == nil {
		t.Error("FetchWithTimeout should have timed out")
	}
}

func TestFetchWithTimeoutClean_Success(t *testing.T) {
	fetcher := func(ctx context.Context) string {
		select {
		case <-time.After(10 * time.Millisecond):
			return "data"
		case <-ctx.Done():
			return ""
		}
	}

	result, err := FetchWithTimeoutClean(fetcher, 1*time.Second)

	if err != nil {
		t.Errorf("FetchWithTimeoutClean error: %v", err)
	}
	if result != "data" {
		t.Errorf("FetchWithTimeoutClean = %q; want \"data\"", result)
	}
}

func TestFetchWithTimeoutClean_Timeout(t *testing.T) {
	fetcher := func(ctx context.Context) string {
		select {
		case <-time.After(1 * time.Second):
			return "slow"
		case <-ctx.Done():
			return ""
		}
	}

	_, err := FetchWithTimeoutClean(fetcher, 50*time.Millisecond)
	if err == nil {
		t.Error("FetchWithTimeoutClean should have timed out")
	}
}

// --- PART 3: Cascading Cancellation ---

func TestProcessPipeline(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	results := ProcessPipeline(ctx)

	if len(results) == 0 {
		t.Error("ProcessPipeline produced 0 results; want at least 1")
	}

	// Verify results are squares
	for _, v := range results {
		// Each value should be a perfect square
		found := false
		for i := 1; i <= 100; i++ {
			if i*i == v {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("ProcessPipeline produced non-square value: %d", v)
		}
	}
}

// --- PART 4: Concurrent Operations ---

func TestFirstSuccess(t *testing.T) {
	fetchers := []func(context.Context) string{
		func(ctx context.Context) string {
			time.Sleep(200 * time.Millisecond)
			return "slow"
		},
		func(ctx context.Context) string {
			time.Sleep(10 * time.Millisecond)
			return "fast"
		},
		func(ctx context.Context) string {
			time.Sleep(100 * time.Millisecond)
			return "medium"
		},
	}

	ctx := context.Background()
	result := FirstSuccess(ctx, fetchers)

	if result != "fast" {
		t.Errorf("FirstSuccess = %q; want \"fast\"", result)
	}
}

func TestAllSuccess(t *testing.T) {
	fetchers := []func(context.Context) string{
		func(ctx context.Context) string { return "a" },
		func(ctx context.Context) string { return "b" },
		func(ctx context.Context) string { return "c" },
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	results := AllSuccess(ctx, fetchers)

	if results == nil {
		t.Fatal("AllSuccess returned nil")
	}
	if len(results) != 3 {
		t.Fatalf("AllSuccess returned %d results; want 3", len(results))
	}
}

// --- PART 5: Context Values ---

func TestWithRequestID(t *testing.T) {
	ctx := context.Background()
	ctx = WithRequestID(ctx, "req-123")

	id := GetRequestID(ctx)
	if id != "req-123" {
		t.Errorf("GetRequestID = %q; want \"req-123\"", id)
	}
}

func TestGetRequestID_Missing(t *testing.T) {
	ctx := context.Background()
	id := GetRequestID(ctx)
	if id != "" {
		t.Errorf("GetRequestID on empty context = %q; want \"\"", id)
	}
}

func TestProcessRequest(t *testing.T) {
	ctx := context.Background()
	ctx = WithRequestID(ctx, "req-456")

	result := ProcessRequest(ctx, "test-data")

	if result == "" {
		t.Error("ProcessRequest returned empty string")
	}
}

// --- Conceptual tests ---

func TestContextConcept(t *testing.T) {
	t.Log("CONCEPT: Context carries cancellation signals, deadlines, and values")
	t.Log("         Cancellation propagates from parent to children automatically")
	t.Log("         Always pass context as first parameter")
	t.Log("         Never store context in a struct")
}
