package concurrency

import (
	"sort"
	"testing"
	"time"
)

// =============================================================================
// Tests for Exercise 3.2: Fan-Out / Fan-In
// Run with: go test -v -run Test09
// =============================================================================

// --- Helper: make a channel from slice ---

func sliceToChan(values []int) <-chan int {
	ch := make(chan int, len(values))
	for _, v := range values {
		ch <- v
	}
	close(ch)
	return ch
}

// --- PART 1: Fan-Out ---

func TestFanOut(t *testing.T) {
	input := sliceToChan([]int{1, 2, 3, 4, 5, 6})

	outputs := FanOut(input, 3)
	if outputs == nil {
		t.Fatal("FanOut returned nil")
	}
	if len(outputs) != 3 {
		t.Fatalf("FanOut returned %d channels; want 3", len(outputs))
	}

	// Collect all values from all output channels
	var all []int
	for i, ch := range outputs {
		for v := range ch {
			all = append(all, v)
			_ = i
		}
	}

	sort.Ints(all)
	expected := []int{1, 2, 3, 4, 5, 6}
	if len(all) != len(expected) {
		t.Fatalf("FanOut produced %d total values; want %d", len(all), len(expected))
	}
	for i, v := range all {
		if v != expected[i] {
			t.Errorf("sorted values[%d] = %d; want %d", i, v, expected[i])
		}
	}
}

func TestFanOutFirstAvailable(t *testing.T) {
	input := sliceToChan([]int{10, 20, 30, 40})

	outputs := FanOutFirstAvailable(input, 2)
	if outputs == nil {
		t.Fatal("FanOutFirstAvailable returned nil")
	}

	var all []int
	for _, ch := range outputs {
		for v := range ch {
			all = append(all, v)
		}
	}

	sort.Ints(all)
	expected := []int{10, 20, 30, 40}
	if len(all) != len(expected) {
		t.Fatalf("produced %d values; want %d", len(all), len(expected))
	}
	for i, v := range all {
		if v != expected[i] {
			t.Errorf("sorted values[%d] = %d; want %d", i, v, expected[i])
		}
	}
}

// --- PART 2: Fan-In ---

func TestFanIn(t *testing.T) {
	ch1 := sliceToChan([]int{1, 3, 5})
	ch2 := sliceToChan([]int{2, 4, 6})

	out := FanIn(ch1, ch2)
	if out == nil {
		t.Fatal("FanIn returned nil")
	}

	var values []int
	for v := range out {
		values = append(values, v)
	}

	sort.Ints(values)
	expected := []int{1, 2, 3, 4, 5, 6}
	if len(values) != len(expected) {
		t.Fatalf("FanIn produced %d values; want %d", len(values), len(expected))
	}
	for i, v := range values {
		if v != expected[i] {
			t.Errorf("sorted values[%d] = %d; want %d", i, v, expected[i])
		}
	}
}

func TestFanInOrdered(t *testing.T) {
	ch1 := sliceToChan([]int{10, 20})
	ch2 := sliceToChan([]int{30, 40})

	out := FanInOrdered(ch1, ch2)
	if out == nil {
		t.Fatal("FanInOrdered returned nil")
	}

	var values []TaggedValue
	for v := range out {
		values = append(values, v)
	}

	if len(values) != 4 {
		t.Fatalf("FanInOrdered produced %d values; want 4", len(values))
	}

	// Check source indices are correct
	for _, tv := range values {
		if tv.SourceIndex == 0 && tv.Value != 10 && tv.Value != 20 {
			t.Errorf("source 0 produced unexpected value %d", tv.Value)
		}
		if tv.SourceIndex == 1 && tv.Value != 30 && tv.Value != 40 {
			t.Errorf("source 1 produced unexpected value %d", tv.Value)
		}
	}
}

// --- PART 3: ParallelProcess ---

func TestParallelProcess(t *testing.T) {
	input := sliceToChan([]int{1, 2, 3, 4, 5})
	double := func(n int) int { return n * 2 }

	out := ParallelProcess(input, 3, double)
	if out == nil {
		t.Fatal("ParallelProcess returned nil")
	}

	var results []int
	for v := range out {
		results = append(results, v)
	}

	sort.Ints(results)
	expected := []int{2, 4, 6, 8, 10}
	if len(results) != len(expected) {
		t.Fatalf("ParallelProcess produced %d values; want %d", len(results), len(expected))
	}
	for i, v := range results {
		if v != expected[i] {
			t.Errorf("sorted results[%d] = %d; want %d", i, v, expected[i])
		}
	}
}

// --- PART 4: ParallelProcessOrdered ---

func TestParallelProcessOrdered(t *testing.T) {
	input := sliceToChan([]int{5, 3, 1, 4, 2})
	square := func(n int) int { return n * n }

	out := ParallelProcessOrdered(input, 3, square)
	if out == nil {
		t.Fatal("ParallelProcessOrdered returned nil")
	}

	var results []int
	for v := range out {
		results = append(results, v)
	}

	// Results should be in input order: 25, 9, 1, 16, 4
	expected := []int{25, 9, 1, 16, 4}
	if len(results) != len(expected) {
		t.Fatalf("produced %d values; want %d", len(results), len(expected))
	}
	for i, v := range results {
		if v != expected[i] {
			t.Errorf("results[%d] = %d; want %d (order-preserving)", i, v, expected[i])
		}
	}
}

// --- PART 5: BoundedFanIn ---

func TestBoundedFanIn(t *testing.T) {
	ch1 := sliceToChan([]int{1, 2, 3})
	ch2 := sliceToChan([]int{4, 5, 6})

	out := BoundedFanIn(2, ch1, ch2)
	if out == nil {
		t.Fatal("BoundedFanIn returned nil")
	}

	var values []int
	for v := range out {
		values = append(values, v)
	}

	sort.Ints(values)
	expected := []int{1, 2, 3, 4, 5, 6}
	if len(values) != len(expected) {
		t.Fatalf("BoundedFanIn produced %d values; want %d", len(values), len(expected))
	}
}

// --- CHALLENGE: MapReduce ---

func TestMapReduce(t *testing.T) {
	input := sliceToChan([]int{1, 2, 3, 4, 5})
	square := func(n int) int { return n * n }
	sum := func(a, b int) int { return a + b }

	result := MapReduce(input, 3, square, sum, 0)
	// 1 + 4 + 9 + 16 + 25 = 55
	if result != 55 {
		t.Errorf("MapReduce (sum of squares) = %d; want 55", result)
	}
}

// --- CHALLENGE: BatchedFanIn ---

func TestBatchedFanIn(t *testing.T) {
	ch := make(chan int, 10)
	for i := 1; i <= 7; i++ {
		ch <- i
	}
	close(ch)

	out := BatchedFanIn(3, 100, ch)
	if out == nil {
		t.Fatal("BatchedFanIn returned nil")
	}

	var batches [][]int
	for batch := range out {
		batches = append(batches, batch)
	}

	// Should have batches of size 3, 3, 1 (or similar)
	totalItems := 0
	for _, batch := range batches {
		totalItems += len(batch)
	}

	if totalItems != 7 {
		t.Errorf("BatchedFanIn produced %d total items; want 7", totalItems)
	}
}

// --- Timeout protection ---

func TestFanIn_Closes(t *testing.T) {
	ch1 := make(chan int)
	close(ch1) // Immediately closed

	ch2 := make(chan int)
	close(ch2)

	out := FanIn(ch1, ch2)
	if out == nil {
		t.Fatal("FanIn returned nil")
	}

	select {
	case _, ok := <-out:
		if ok {
			// Value received, that's fine
		}
		// Channel closed, expected
	case <-time.After(time.Second):
		t.Error("FanIn output channel did not close within 1 second")
	}
}
