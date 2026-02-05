package concurrency

import (
	"context"
	"runtime"
	"time"
)

// =============================================================================
// EXERCISE 4.3: Goroutine Leaks
// =============================================================================
//
// A goroutine leak occurs when goroutines are created but never terminate.
// This causes memory growth and can eventually crash your program.
//
// KEY CONCEPTS:
// - Goroutines are NOT garbage collected - they must exit on their own
// - Common causes: blocked on channel, waiting for lock, infinite loop
// - Detection: runtime.NumGoroutine(), pprof, goleak library
// - Prevention: always ensure goroutines have an exit path
// - Solution: context cancellation, done channels, timeouts
//
// =============================================================================

// =============================================================================
// PART 1: Detecting Goroutine Leaks
// =============================================================================

// CountGoroutines returns the current number of goroutines.
// Use this to detect leaks in tests.
func CountGoroutines() int {
	return runtime.NumGoroutine()
}

// LeakyFunction creates goroutines that never exit.
// QUESTION: Why don't these goroutines terminate?
func LeakyFunction() {
	ch := make(chan int)

	go func() {
		val := <-ch // LEAK: blocks forever, no sender
		_ = val
	}()

	go func() {
		ch <- 42 // LEAK: blocks forever, no receiver
	}()

	// Function returns, but goroutines are stuck
}

// FixedFunction should not leak goroutines.
//
// TODO: Fix so goroutines can exit
// HINT: Use context or done channel
func FixedFunction() {
	// YOUR CODE HERE
}

// =============================================================================
// PART 2: Leak from Abandoned Channel
// =============================================================================

// FetchFirst gets the first result from multiple fetchers.
// The other fetchers leak because their results are never consumed!
func FetchFirst(fetchers []func() string) string {
	results := make(chan string)

	for _, fetch := range fetchers {
		go func(f func() string) {
			results <- f() // LEAK: losers block here forever
		}(fetch)
	}

	return <-results // Only consume first result
}

// FetchFirstSafe gets the first result without leaking goroutines.
//
// TODO: Fix the leak
// HINT: Buffered channel OR context cancellation
func FetchFirstSafe(fetchers []func() string) string {
	// YOUR CODE HERE
	return ""
}

// FetchFirstWithCancel gets first result, properly cancelling others.
//
// TODO: Implement using context to cancel slow fetchers
// This is the preferred pattern for production code
func FetchFirstWithCancel(ctx context.Context, fetchers []func(context.Context) string) string {
	// YOUR CODE HERE
	return ""
}

// =============================================================================
// PART 3: Leak from Forgotten Done Channel
// =============================================================================

// WorkerLeak starts a worker that leaks.
// QUESTION: How does the worker know when to stop?
func WorkerLeak() <-chan int {
	out := make(chan int)

	go func() {
		i := 0
		for {
			out <- i // LEAK: if receiver stops reading, blocks forever
			i++
		}
	}()

	return out
}

// WorkerWithDone starts a worker that can be stopped.
//
// TODO: Implement worker that stops when done channel is closed
func WorkerWithDone(done <-chan struct{}) <-chan int {
	// YOUR CODE HERE
	return nil
}

// WorkerWithContext starts a worker that respects context.
//
// TODO: Implement worker that stops when context is cancelled
func WorkerWithContext(ctx context.Context) <-chan int {
	// YOUR CODE HERE
	return nil
}

// =============================================================================
// PART 4: Leak from Unbounded Producer
// =============================================================================

// GenerateForever produces values forever.
// If consumer stops, producer leaks.
func GenerateForever(values []int) <-chan int {
	out := make(chan int)

	go func() {
		for {
			for _, v := range values {
				out <- v // LEAK: blocks if no consumer
			}
		}
	}()

	return out
}

// GenerateWithLimit produces values until limit or cancellation.
//
// TODO: Implement bounded producer
// Accept context and maximum count
func GenerateWithLimit(ctx context.Context, values []int, maxCount int) <-chan int {
	// YOUR CODE HERE
	return nil
}

// =============================================================================
// PART 5: Leak from Timer/Ticker
// =============================================================================

// LeakyTimer demonstrates resource waste with time.After.
//
// CLARIFICATION: time.After doesn't actually create a goroutine that leaks
// forever. It creates an internal timer that will fire after the duration,
// then be garbage collected. However:
// 1. The timer resources (memory, runtime timer) aren't freed until it fires
// 2. In hot loops, this can create thousands of pending timers
// 3. This is resource waste, not a permanent leak
//
// Use time.NewTimer with explicit Stop() for better resource management.
func LeakyTimer() string {
	ch := make(chan string)

	go func() {
		// Simulate slow work
		// time.Sleep(200 * time.Millisecond)
		ch <- "result"
	}()

	select {
	case result := <-ch:
		return result
	case <-time.After(100 * time.Millisecond):
		return "timeout"
		// RESOURCE WASTE: timer resources held until it would have fired
		// In loops, this accumulates. Use time.NewTimer + Stop() instead.
	}
}

// SafeTimer uses time.NewTimer for proper cleanup.
//
// TODO: Implement using time.NewTimer with proper Stop()
func SafeTimer() string {
	// YOUR CODE HERE
	return ""
}

// LeakyTicker creates a ticker that leaks if not stopped.
func LeakyTicker() int {
	ticker := time.NewTicker(10 * time.Millisecond)
	// Forgot to call ticker.Stop()!

	count := 0
	timeout := time.After(50 * time.Millisecond)

	for {
		select {
		case <-ticker.C:
			count++
		case <-timeout:
			return count
			// LEAK: ticker goroutine continues forever
		}
	}
}

// SafeTicker properly stops the ticker.
//
// TODO: Implement with proper ticker cleanup
func SafeTicker() int {
	// YOUR CODE HERE
	return 0
}

// =============================================================================
// PART 6: Testing for Leaks
// =============================================================================

// RunWithLeakCheck runs a function and checks for goroutine leaks.
// Returns true if no leak detected.
//
// TODO: Implement leak detection helper
// 1. Count goroutines before running fn
// 2. Run fn
// 3. Wait briefly for goroutines to exit
// 4. Count goroutines after
// 5. Return true if count is same or less
func RunWithLeakCheck(fn func()) bool {
	// YOUR CODE HERE
	return false
}

// =============================================================================
// CHALLENGE: Fix a complex leak scenario
// =============================================================================

// Pipeline creates a data processing pipeline that leaks.
// Find all the leaks and fix them!
func LeakyPipeline(input []int) []int {
	// Stage 1: Generate
	gen := make(chan int)
	go func() {
		for _, v := range input {
			gen <- v
		}
		// LEAK 1: forgot to close(gen)
	}()

	// Stage 2: Transform
	transform := make(chan int)
	go func() {
		for v := range gen {
			transform <- v * 2
		}
		// LEAK 2: forgot to close(transform)
	}()

	// Stage 3: Collect first N
	var results []int
	for i := 0; i < 3; i++ { // Only take 3 values
		results = append(results, <-transform)
	}
	// LEAK 3: remaining pipeline stages blocked

	return results
}

// SafePipeline should not leak for any input size.
//
// TODO: Fix all leaks using context cancellation
func SafePipeline(ctx context.Context, input []int, maxResults int) []int {
	// YOUR CODE HERE
	return nil
}

// Ensure imports are used
var _ = context.Background
var _ = runtime.NumGoroutine
var _ = time.Second
