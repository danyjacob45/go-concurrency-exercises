package concurrency

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// =============================================================================
// Tests for Exercise 5.3: singleflight for Request Deduplication
// Run with: go test -v -run Test18
// =============================================================================

// --- PART 1: SingleFlight.Do ---

func TestSingleFlight_Do_Basic(t *testing.T) {
	sf := NewSingleFlight()
	if sf == nil {
		t.Fatal("NewSingleFlight returned nil")
	}

	val, err, shared := sf.Do("key", func() (any, error) {
		return "result", nil
	})

	if err != nil {
		t.Errorf("Do() error = %v", err)
	}
	if val != "result" {
		t.Errorf("Do() value = %v; want \"result\"", val)
	}
	if shared {
		t.Error("Do() shared = true; want false (only one caller)")
	}
}

func TestSingleFlight_Do_Deduplication(t *testing.T) {
	sf := NewSingleFlight()
	if sf == nil {
		t.Fatal("NewSingleFlight returned nil")
	}

	var callCount int64

	var wg sync.WaitGroup
	n := 10

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sf.Do("key", func() (any, error) {
				atomic.AddInt64(&callCount, 1)
				time.Sleep(50 * time.Millisecond) // Slow operation
				return "shared-result", nil
			})
		}()
	}

	wg.Wait()

	// Function should only be called once
	count := atomic.LoadInt64(&callCount)
	if count != 1 {
		t.Errorf("function called %d times; want 1 (deduplication failed)", count)
	}
}

func TestSingleFlight_Do_DifferentKeys(t *testing.T) {
	sf := NewSingleFlight()
	if sf == nil {
		t.Fatal("NewSingleFlight returned nil")
	}

	var callCount int64
	var wg sync.WaitGroup

	// Different keys should not be deduplicated
	for i := 0; i < 3; i++ {
		key := string(rune('a' + i))
		wg.Add(1)
		go func(k string) {
			defer wg.Done()
			sf.Do(k, func() (any, error) {
				atomic.AddInt64(&callCount, 1)
				return k, nil
			})
		}(key)
	}

	wg.Wait()

	count := atomic.LoadInt64(&callCount)
	if count != 3 {
		t.Errorf("function called %d times; want 3 (different keys)", count)
	}
}

// --- Forget ---

func TestSingleFlight_Forget(t *testing.T) {
	sf := NewSingleFlight()
	if sf == nil {
		t.Fatal("NewSingleFlight returned nil")
	}

	// First call
	sf.Do("key", func() (any, error) {
		return "first", nil
	})

	sf.Forget("key")

	// After Forget, new call should execute
	val, _, _ := sf.Do("key", func() (any, error) {
		return "second", nil
	})

	if val != "second" {
		t.Errorf("after Forget, Do() = %v; want \"second\"", val)
	}
}

// --- PART 2: DoChan ---

func TestSingleFlight_DoChan(t *testing.T) {
	sf := NewSingleFlight()
	if sf == nil {
		t.Fatal("NewSingleFlight returned nil")
	}

	ch := sf.DoChan("key", func() (any, error) {
		time.Sleep(10 * time.Millisecond)
		return "async-result", nil
	})

	if ch == nil {
		t.Fatal("DoChan returned nil channel")
	}

	select {
	case result := <-ch:
		if result.Val != "async-result" {
			t.Errorf("DoChan result = %v; want \"async-result\"", result.Val)
		}
		if result.Err != nil {
			t.Errorf("DoChan error = %v; want nil", result.Err)
		}
	case <-time.After(time.Second):
		t.Fatal("DoChan timed out")
	}
}

// --- PART 3: CachedFetcher ---

func TestCachedFetcher_Basic(t *testing.T) {
	cf := NewCachedFetcher()
	if cf == nil {
		t.Fatal("NewCachedFetcher returned nil")
	}

	var fetchCount int64
	fetcher := func(key string) (string, error) {
		atomic.AddInt64(&fetchCount, 1)
		return "value-" + key, nil
	}

	// First fetch - should call fetcher
	v1, err := cf.Fetch("key1", fetcher, time.Minute)
	if err != nil {
		t.Errorf("Fetch error: %v", err)
	}
	if v1 != "value-key1" {
		t.Errorf("Fetch = %q; want \"value-key1\"", v1)
	}

	// Second fetch - should use cache
	v2, err := cf.Fetch("key1", fetcher, time.Minute)
	if err != nil {
		t.Errorf("Fetch error: %v", err)
	}
	if v2 != "value-key1" {
		t.Errorf("cached Fetch = %q; want \"value-key1\"", v2)
	}

	if atomic.LoadInt64(&fetchCount) != 1 {
		t.Errorf("fetcher called %d times; want 1 (should use cache)", fetchCount)
	}
}

func TestCachedFetcher_Deduplication(t *testing.T) {
	cf := NewCachedFetcher()
	if cf == nil {
		t.Fatal("NewCachedFetcher returned nil")
	}

	var fetchCount int64
	fetcher := func(key string) (string, error) {
		atomic.AddInt64(&fetchCount, 1)
		time.Sleep(50 * time.Millisecond) // Slow fetch
		return "val", nil
	}

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cf.Fetch("same-key", fetcher, time.Minute)
		}()
	}

	wg.Wait()

	count := atomic.LoadInt64(&fetchCount)
	if count != 1 {
		t.Errorf("fetcher called %d times; want 1 (singleflight deduplication)", count)
	}
}

// --- PART 4: DoWithTimeout ---

func TestSingleFlight_DoWithTimeout_Success(t *testing.T) {
	sf := NewSingleFlight()
	if sf == nil {
		t.Fatal("NewSingleFlight returned nil")
	}

	val, err, _ := sf.DoWithTimeout("key", func() (any, error) {
		return "fast", nil
	}, time.Second)

	if err != nil {
		t.Errorf("DoWithTimeout error: %v", err)
	}
	if val != "fast" {
		t.Errorf("DoWithTimeout = %v; want \"fast\"", val)
	}
}

func TestSingleFlight_DoWithTimeout_Timeout(t *testing.T) {
	sf := NewSingleFlight()
	if sf == nil {
		t.Fatal("NewSingleFlight returned nil")
	}

	_, err, _ := sf.DoWithTimeout("key", func() (any, error) {
		time.Sleep(1 * time.Second) // Slow
		return "slow", nil
	}, 50*time.Millisecond)

	if err == nil {
		t.Error("DoWithTimeout should have timed out")
	}
}

func TestSingleFlight_DoWithContext_Cancelled(t *testing.T) {
	sf := NewSingleFlight()
	if sf == nil {
		t.Fatal("NewSingleFlight returned nil")
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err, _ := sf.DoWithContext(ctx, "key", func() (any, error) {
		time.Sleep(100 * time.Millisecond)
		return "val", nil
	})

	if err == nil {
		t.Error("DoWithContext should return error for cancelled context")
	}
}

// --- PART 5: DoFresh ---

func TestSingleFlight_DoFresh(t *testing.T) {
	sf := NewSingleFlight()
	if sf == nil {
		t.Fatal("NewSingleFlight returned nil")
	}

	val, err := sf.DoFresh("key", func() (any, error) {
		return "fresh", nil
	})

	if err != nil {
		t.Errorf("DoFresh error: %v", err)
	}
	if val != "fresh" {
		t.Errorf("DoFresh = %v; want \"fresh\"", val)
	}
}

// --- PART 6: RequestCoalescer ---

func TestRequestCoalescer(t *testing.T) {
	batchFn := func(keys []string) (map[string]any, error) {
		result := make(map[string]any)
		for _, k := range keys {
			result[k] = "val-" + k
		}
		return result, nil
	}

	rc := NewRequestCoalescer(50*time.Millisecond, batchFn)
	if rc == nil {
		t.Fatal("NewRequestCoalescer returned nil")
	}

	var wg sync.WaitGroup
	results := make(map[string]any)
	var mu sync.Mutex

	for _, key := range []string{"a", "b", "c"} {
		wg.Add(1)
		go func(k string) {
			defer wg.Done()
			val, err := rc.Get(k)
			if err != nil {
				t.Errorf("Get(%q) error: %v", k, err)
				return
			}
			mu.Lock()
			results[k] = val
			mu.Unlock()
		}(key)
	}

	wg.Wait()

	for _, key := range []string{"a", "b", "c"} {
		expected := "val-" + key
		if results[key] != expected {
			t.Errorf("results[%q] = %v; want %q", key, results[key], expected)
		}
	}
}

// --- Conceptual tests ---

func TestSingleFlightConcept(t *testing.T) {
	t.Log("CONCEPT: singleflight prevents thundering herd / cache stampede")
	t.Log("         Only one execution per key at a time")
	t.Log("         All callers share the result of the single execution")
	t.Log("         Combine with cache for the classic cache+singleflight pattern")
}
