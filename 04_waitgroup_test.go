package concurrency

import (
	"errors"
	"fmt"
	"sync"
	"testing"
)

// =============================================================================
// Tests for Exercise 2.1: sync.WaitGroup
// Run with: go test -v -run Test04
// =============================================================================

// --- PART 1: Basic WaitGroup Usage ---

func TestConcurrentSum(t *testing.T) {
	double := func(n int) int { return n * 2 }

	result := ConcurrentSum([]int{1, 2, 3, 4, 5}, double)
	expected := 2 + 4 + 6 + 8 + 10 // = 30

	if result != expected {
		t.Errorf("ConcurrentSum([1..5], double) = %d; want %d", result, expected)
	}
}

func TestConcurrentSum_Empty(t *testing.T) {
	result := ConcurrentSum([]int{}, func(n int) int { return n })
	if result != 0 {
		t.Errorf("ConcurrentSum([], identity) = %d; want 0", result)
	}
}

func TestConcurrentSum_Identity(t *testing.T) {
	result := ConcurrentSum([]int{10, 20, 30}, func(n int) int { return n })
	if result != 60 {
		t.Errorf("ConcurrentSum([10,20,30], identity) = %d; want 60", result)
	}
}

func TestFetchAll(t *testing.T) {
	fetcher := func(url string) string { return "data-from-" + url }
	urls := []string{"url1", "url2", "url3"}

	results := FetchAll(urls, fetcher)

	if results == nil {
		t.Fatal("FetchAll returned nil")
	}

	if len(results) != 3 {
		t.Fatalf("FetchAll returned %d results; want 3", len(results))
	}

	for _, url := range urls {
		expected := "data-from-" + url
		if results[url] != expected {
			t.Errorf("results[%q] = %q; want %q", url, results[url], expected)
		}
	}
}

func TestFetchAll_Empty(t *testing.T) {
	results := FetchAll([]string{}, func(url string) string { return url })
	if results == nil {
		t.Fatal("FetchAll returned nil for empty input")
	}
	if len(results) != 0 {
		t.Errorf("FetchAll returned %d results for empty input; want 0", len(results))
	}
}

// --- PART 2: WaitGroup Gotchas ---

func TestProcessBatches(t *testing.T) {
	process := func(n int) int { return n * 2 }
	items := []int{1, 2, 3, 4, 5, 6, 7}
	batchSize := 3

	results := ProcessBatches(items, batchSize, process)

	if results == nil {
		t.Fatal("ProcessBatches returned nil")
	}

	if len(results) != len(items) {
		t.Fatalf("ProcessBatches returned %d results; want %d", len(results), len(items))
	}

	for i, item := range items {
		expected := item * 2
		if results[i] != expected {
			t.Errorf("results[%d] = %d; want %d", i, results[i], expected)
		}
	}
}

func TestWaitGroupReuse(t *testing.T) {
	n := 3
	results := WaitGroupReuse(n)

	if results == nil {
		t.Fatal("WaitGroupReuse returned nil")
	}

	if len(results) != 2*n {
		t.Fatalf("WaitGroupReuse(%d) returned %d results; want %d", n, len(results), 2*n)
	}

	round1Count := 0
	round2Count := 0
	for _, r := range results {
		switch r {
		case "round1":
			round1Count++
		case "round2":
			round2Count++
		default:
			t.Errorf("unexpected result: %q", r)
		}
	}

	if round1Count != n {
		t.Errorf("round1 count = %d; want %d", round1Count, n)
	}
	if round2Count != n {
		t.Errorf("round2 count = %d; want %d", round2Count, n)
	}
}

// --- PART 3: WaitGroup with Errors ---

func TestFirstError_AllSucceed(t *testing.T) {
	fns := []func() error{
		func() error { return nil },
		func() error { return nil },
		func() error { return nil },
	}

	err := FirstError(fns)
	if err != nil {
		t.Errorf("FirstError with all success = %v; want nil", err)
	}
}

func TestFirstError_OneError(t *testing.T) {
	expected := errors.New("something failed")
	fns := []func() error{
		func() error { return nil },
		func() error { return expected },
		func() error { return nil },
	}

	err := FirstError(fns)
	if err == nil {
		t.Error("FirstError with one failing = nil; want error")
	}
}

func TestFirstError_MultipleErrors(t *testing.T) {
	fns := []func() error{
		func() error { return errors.New("error1") },
		func() error { return errors.New("error2") },
		func() error { return errors.New("error3") },
	}

	err := FirstError(fns)
	if err == nil {
		t.Error("FirstError with all failing = nil; want error")
	}
}

func TestAllErrors_AllSucceed(t *testing.T) {
	fns := []func() error{
		func() error { return nil },
		func() error { return nil },
	}

	errs := AllErrors(fns)
	if len(errs) != 0 {
		t.Errorf("AllErrors with all success returned %d errors; want 0", len(errs))
	}
}

func TestAllErrors_SomeErrors(t *testing.T) {
	fns := []func() error{
		func() error { return nil },
		func() error { return errors.New("err1") },
		func() error { return nil },
		func() error { return errors.New("err2") },
	}

	errs := AllErrors(fns)
	if len(errs) != 2 {
		t.Errorf("AllErrors returned %d errors; want 2", len(errs))
	}
}

// --- CHALLENGE: ParallelMap ---

func TestParallelMap_Success(t *testing.T) {
	fn := func(n int) (int, error) { return n * n, nil }

	results, err := ParallelMap([]int{1, 2, 3, 4}, fn)
	if err != nil {
		t.Fatalf("ParallelMap returned error: %v", err)
	}

	expected := []int{1, 4, 9, 16}
	if len(results) != len(expected) {
		t.Fatalf("ParallelMap returned %d results; want %d", len(results), len(expected))
	}

	for i, v := range results {
		if v != expected[i] {
			t.Errorf("results[%d] = %d; want %d", i, v, expected[i])
		}
	}
}

func TestParallelMap_Error(t *testing.T) {
	fn := func(n int) (int, error) {
		if n == 3 {
			return 0, fmt.Errorf("error on %d", n)
		}
		return n * n, nil
	}

	_, err := ParallelMap([]int{1, 2, 3, 4}, fn)
	if err == nil {
		t.Error("ParallelMap with error returned nil; want error")
	}
}

// --- Concurrency safety test ---

func TestConcurrentSum_RaceFree(t *testing.T) {
	// Run many times to increase chance of detecting races
	for i := 0; i < 10; i++ {
		nums := make([]int, 100)
		for j := range nums {
			nums[j] = 1
		}
		result := ConcurrentSum(nums, func(n int) int { return n })
		if result != 100 {
			t.Errorf("iteration %d: ConcurrentSum = %d; want 100", i, result)
		}
	}
}

func TestFetchAll_Concurrent(t *testing.T) {
	var mu sync.Mutex
	callCount := 0

	fetcher := func(url string) string {
		mu.Lock()
		callCount++
		mu.Unlock()
		return "result"
	}

	urls := make([]string, 50)
	for i := range urls {
		urls[i] = fmt.Sprintf("url%d", i)
	}

	results := FetchAll(urls, fetcher)
	if len(results) != 50 {
		t.Errorf("FetchAll returned %d results; want 50", len(results))
	}

	mu.Lock()
	if callCount != 50 {
		t.Errorf("fetcher called %d times; want 50", callCount)
	}
	mu.Unlock()
}
