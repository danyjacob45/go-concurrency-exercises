package concurrency

import (
	"sync/atomic"
	"testing"
	"time"
)

// =============================================================================
// Tests for Exercise 3.1: Worker Pool Pattern
// Run with: go test -v -run Test08
// =============================================================================

// --- PART 1: Basic Worker Pool ---

func TestWorkerPool_Basic(t *testing.T) {
	process := func(j Job) JobResult {
		return JobResult{JobID: j.ID, Output: j.Payload * 2}
	}

	wp := NewWorkerPool(3, process)
	if wp == nil {
		t.Fatal("NewWorkerPool returned nil")
	}
	defer wp.Shutdown()

	// Submit jobs
	for i := 0; i < 5; i++ {
		ok := wp.Submit(Job{ID: i, Payload: i + 1})
		if !ok {
			t.Errorf("Submit job %d returned false", i)
		}
	}

	// Collect results
	results := make(map[int]int)
	resultsCh := wp.Results()
	if resultsCh == nil {
		t.Fatal("Results() returned nil channel")
	}

	for i := 0; i < 5; i++ {
		select {
		case r := <-resultsCh:
			results[r.JobID] = r.Output
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for result")
		}
	}

	// Verify results
	for i := 0; i < 5; i++ {
		expected := (i + 1) * 2
		if results[i] != expected {
			t.Errorf("result for job %d = %d; want %d", i, results[i], expected)
		}
	}
}

func TestWorkerPool_Shutdown(t *testing.T) {
	var processed int64

	process := func(j Job) JobResult {
		atomic.AddInt64(&processed, 1)
		time.Sleep(10 * time.Millisecond)
		return JobResult{JobID: j.ID, Output: j.Payload}
	}

	wp := NewWorkerPool(2, process)
	if wp == nil {
		t.Fatal("NewWorkerPool returned nil")
	}

	for i := 0; i < 5; i++ {
		wp.Submit(Job{ID: i, Payload: i})
	}

	wp.Shutdown()

	// After shutdown, all submitted jobs should be processed
	if atomic.LoadInt64(&processed) != 5 {
		t.Errorf("processed %d jobs; want 5", processed)
	}
}

func TestWorkerPool_SubmitAfterShutdown(t *testing.T) {
	process := func(j Job) JobResult {
		return JobResult{JobID: j.ID, Output: j.Payload}
	}

	wp := NewWorkerPool(2, process)
	if wp == nil {
		t.Fatal("NewWorkerPool returned nil")
	}

	wp.Shutdown()

	ok := wp.Submit(Job{ID: 1, Payload: 1})
	if ok {
		t.Error("Submit after Shutdown returned true; want false")
	}
}

func TestWorkerPool_ManyJobs(t *testing.T) {
	process := func(j Job) JobResult {
		return JobResult{JobID: j.ID, Output: j.Payload * j.Payload}
	}

	wp := NewWorkerPool(4, process)
	if wp == nil {
		t.Fatal("NewWorkerPool returned nil")
	}

	n := 100
	for i := 0; i < n; i++ {
		wp.Submit(Job{ID: i, Payload: i})
	}

	results := make(map[int]int)
	resultsCh := wp.Results()

	for i := 0; i < n; i++ {
		select {
		case r := <-resultsCh:
			results[r.JobID] = r.Output
		case <-time.After(5 * time.Second):
			t.Fatalf("timed out after receiving %d results", i)
		}
	}

	wp.Shutdown()

	if len(results) != n {
		t.Errorf("got %d results; want %d", len(results), n)
	}
}

// --- Conceptual tests ---

func TestWorkerPoolConcept(t *testing.T) {
	t.Log("CONCEPT: Worker pools limit goroutine count to prevent resource exhaustion")
	t.Log("         Workers pull from a shared job channel (fan-out)")
	t.Log("         Results are pushed to a shared result channel (fan-in)")
	t.Log("         Graceful shutdown: close job channel, wait for workers, close results")
}
