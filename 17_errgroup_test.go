package concurrency

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
)

// =============================================================================
// Tests for Exercise 5.2: errgroup for Goroutine Management
// Run with: go test -v -run Test17
// =============================================================================

// --- PART 1: Custom ErrGroup ---

func TestErrGroup_AllSuccess(t *testing.T) {
	var g ErrGroup

	for i := 0; i < 5; i++ {
		g.Go(func() error {
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		t.Errorf("Wait() = %v; want nil", err)
	}
}

func TestErrGroup_FirstError(t *testing.T) {
	var g ErrGroup

	g.Go(func() error { return nil })
	g.Go(func() error { return errors.New("oops") })
	g.Go(func() error { return nil })

	if err := g.Wait(); err == nil {
		t.Error("Wait() = nil; want error")
	}
}

func TestErrGroup_AllErrors(t *testing.T) {
	var g ErrGroup

	g.Go(func() error { return errors.New("err1") })
	g.Go(func() error { return errors.New("err2") })

	if err := g.Wait(); err == nil {
		t.Error("Wait() = nil; want error")
	}
}

// --- PART 2: ErrGroupWithContext ---

func TestErrGroupWithContext_Cancellation(t *testing.T) {
	g, ctx := WithContext(context.Background())
	if g == nil {
		t.Fatal("WithContext returned nil group")
	}

	g.Go(func() error {
		return errors.New("fail fast")
	})

	g.Go(func() error {
		// This should see context cancelled
		<-ctx.Done()
		return ctx.Err()
	})

	err := g.Wait()
	if err == nil {
		t.Error("Wait() = nil; want error")
	}
}

// --- PART 3: FetchAllURLs ---

func TestFetchAllURLs_Success(t *testing.T) {
	fetcher := func(ctx context.Context, url string) (string, error) {
		return "data-" + url, nil
	}

	results, err := FetchAllURLs(context.Background(), []string{"a", "b", "c"}, fetcher)

	if err != nil {
		t.Errorf("FetchAllURLs error: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("got %d results; want 3", len(results))
	}
}

func TestFetchAllURLs_Error(t *testing.T) {
	fetcher := func(ctx context.Context, url string) (string, error) {
		if url == "bad" {
			return "", errors.New("fetch failed")
		}
		return "ok", nil
	}

	_, err := FetchAllURLs(context.Background(), []string{"good", "bad", "good"}, fetcher)
	if err == nil {
		t.Error("FetchAllURLs should return error when a fetch fails")
	}
}

func TestFetchAllURLsPartial(t *testing.T) {
	fetcher := func(ctx context.Context, url string) (string, error) {
		if url == "bad" {
			return "", errors.New("failed")
		}
		return "data-" + url, nil
	}

	results, err := FetchAllURLsPartial(context.Background(), []string{"a", "bad", "c"}, fetcher)

	// Should have partial results
	if results == nil {
		t.Fatal("FetchAllURLsPartial returned nil results")
	}
	if len(results) < 2 {
		t.Errorf("got %d partial results; want at least 2", len(results))
	}
	if err == nil {
		t.Error("FetchAllURLsPartial should return error for failed fetch")
	}
}

// --- PART 4: LimitedErrGroup ---

func TestLimitedErrGroup_Limit(t *testing.T) {
	g, _ := NewLimitedErrGroup(context.Background(), 2)
	if g == nil {
		t.Fatal("NewLimitedErrGroup returned nil")
	}

	var mu sync.Mutex
	var results []int

	for i := 0; i < 10; i++ {
		val := i
		g.Go(func() error {
			mu.Lock()
			results = append(results, val)
			mu.Unlock()
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		t.Errorf("Wait() = %v; want nil", err)
	}

	if len(results) != 10 {
		t.Errorf("processed %d items; want 10", len(results))
	}
}

func TestProcessWithLimit(t *testing.T) {
	process := func(ctx context.Context, n int) (int, error) {
		return n * n, nil
	}

	results, err := ProcessWithLimit(context.Background(), []int{1, 2, 3, 4, 5}, 2, process)

	if err != nil {
		t.Errorf("ProcessWithLimit error: %v", err)
	}
	if len(results) != 5 {
		t.Fatalf("got %d results; want 5", len(results))
	}

	expected := []int{1, 4, 9, 16, 25}
	for i, v := range results {
		if v != expected[i] {
			t.Errorf("results[%d] = %d; want %d", i, v, expected[i])
		}
	}
}

// --- PART 5: MultiError ---

func TestMultiError(t *testing.T) {
	me := &MultiError{}

	if me.HasErrors() {
		t.Error("new MultiError HasErrors() = true; want false")
	}

	me.Add(errors.New("err1"))
	me.Add(errors.New("err2"))

	if !me.HasErrors() {
		t.Error("MultiError HasErrors() = false after adding errors")
	}

	if len(me.Errors()) != 2 {
		t.Errorf("Errors() length = %d; want 2", len(me.Errors()))
	}

	if me.Error() == "" {
		t.Error("Error() returned empty string")
	}
}

func TestErrGroupMulti(t *testing.T) {
	var g ErrGroupMulti

	g.Go(func() error { return errors.New("err1") })
	g.Go(func() error { return nil })
	g.Go(func() error { return errors.New("err2") })

	err := g.Wait()
	if err == nil {
		t.Fatal("Wait() = nil; want MultiError")
	}

	me, ok := err.(*MultiError)
	if !ok {
		t.Fatalf("error type = %T; want *MultiError", err)
	}

	if len(me.Errors()) != 2 {
		t.Errorf("MultiError has %d errors; want 2", len(me.Errors()))
	}
}

// --- PART 6: Pipeline ---

func TestRunErrGroupPipeline(t *testing.T) {
	input := make(chan int, 5)
	for i := 1; i <= 5; i++ {
		input <- i
	}
	close(input)

	double := func(ctx context.Context, in <-chan int, out chan<- int) error {
		for v := range in {
			select {
			case out <- v * 2:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		return nil
	}

	addOne := func(ctx context.Context, in <-chan int, out chan<- int) error {
		for v := range in {
			select {
			case out <- v + 1:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		return nil
	}

	output, wait := RunErrGroupPipeline(context.Background(), input, double, addOne)
	if output == nil {
		t.Fatal("RunErrGroupPipeline returned nil output channel")
	}

	var results []int
	for v := range output {
		results = append(results, v)
	}

	if err := wait(); err != nil {
		t.Errorf("pipeline error: %v", err)
	}

	// Input: 1,2,3,4,5 -> double: 2,4,6,8,10 -> addOne: 3,5,7,9,11
	if len(results) != 5 {
		t.Fatalf("pipeline produced %d values; want 5", len(results))
	}
}

// --- CHALLENGE: RunWithRetry ---

func TestRunWithRetry(t *testing.T) {
	callCounts := make(map[int]int)
	var mu sync.Mutex

	tasks := []RetryTask{
		{
			ID:      1,
			Retries: 3,
			Fn: func(ctx context.Context) error {
				mu.Lock()
				callCounts[1]++
				count := callCounts[1]
				mu.Unlock()
				if count < 3 {
					return fmt.Errorf("not yet")
				}
				return nil
			},
		},
		{
			ID:      2,
			Retries: 0,
			Fn: func(ctx context.Context) error {
				return nil
			},
		},
	}

	results := RunWithRetry(context.Background(), tasks, 2)
	if results == nil {
		t.Fatal("RunWithRetry returned nil")
	}

	// Task 2 should succeed
	if results[2] != nil {
		t.Errorf("task 2 error = %v; want nil", results[2])
	}
}

// --- Conceptual tests ---

func TestErrGroupConcept(t *testing.T) {
	t.Log("CONCEPT: errgroup manages goroutine lifecycle and error collection")
	t.Log("         Go() launches, Wait() collects first error")
	t.Log("         WithContext() cancels remaining goroutines on first error")
	t.Log("         SetLimit() controls concurrency")
}
