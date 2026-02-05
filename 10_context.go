package concurrency

import (
	"context"
	"time"
)

// =============================================================================
// EXERCISE 3.3: Context and Cancellation
// =============================================================================
//
// context.Context provides cancellation signals, deadlines, and request-scoped
// values across API boundaries and goroutines.
//
// KEY CONCEPTS:
// - context.Background() - root context, never cancelled
// - context.WithCancel(parent) - returns ctx and cancel func
// - context.WithTimeout(parent, duration) - auto-cancels after duration
// - context.WithDeadline(parent, time) - auto-cancels at specific time
// - ctx.Done() - channel that closes when context is cancelled
// - ctx.Err() - returns why context was cancelled (Canceled or DeadlineExceeded)
// - Cancellation propagates from parent to children, not vice versa
//
// GO 1.20+ ADDITIONS:
// - context.WithCancelCause(parent) - cancel with a custom error cause
// - context.Cause(ctx) - retrieve the cause passed to cancel
//
// GO 1.21+ ADDITIONS:
// - context.WithoutCancel(parent) - derive context that ignores parent cancellation
// - context.AfterFunc(ctx, f) - run f when context is cancelled
//
// =============================================================================

// =============================================================================
// PART 1: Basic Cancellation
// =============================================================================

// DoWorkWithCancel does work until cancelled.
// Returns the amount of work completed.
//
// TODO: Implement to:
// 1. Loop doing "work" (increment counter, sleep 10ms)
// 2. Check ctx.Done() each iteration
// 3. Return count when cancelled
//
// QUESTION: Why check ctx.Done() in a select instead of just <-ctx.Done()?
func DoWorkWithCancel(ctx context.Context) int {
	// YOUR CODE HERE
	return 0
}

// SearchWithCancel searches for target in a slow data source.
// Returns (result, true) if found, ("", false) if cancelled or not found.
//
// TODO: Implement to:
// 1. Iterate through dataSource (each item takes 50ms to "fetch")
// 2. Return immediately if ctx is cancelled
// 3. Return result if found
func SearchWithCancel(ctx context.Context, dataSource []string, target string) (string, bool) {
	// YOUR CODE HERE
	return "", false
}

// =============================================================================
// PART 2: Timeouts
// =============================================================================

// FetchWithTimeout fetches data with a timeout.
// The fetcher function simulates a slow operation.
//
// TODO: Implement to:
// 1. Create a context with timeout
// 2. Run fetcher in a goroutine
// 3. Return result if fetcher completes in time
// 4. Return error if timeout
//
// QUESTION: What happens to the goroutine if timeout occurs? (Goroutine leak!)
func FetchWithTimeout(fetcher func() string, timeout time.Duration) (string, error) {
	// YOUR CODE HERE
	return "", nil
}

// FetchWithTimeoutClean is like above but avoids goroutine leaks.
//
// TODO: Pass context to fetcher so it can check for cancellation
// The fetcher should periodically check ctx.Done() and return early
func FetchWithTimeoutClean(fetcher func(ctx context.Context) string, timeout time.Duration) (string, error) {
	// YOUR CODE HERE
	return "", nil
}

// =============================================================================
// PART 3: Cascading Cancellation
// =============================================================================

// ProcessPipeline runs a three-stage pipeline that respects cancellation.
// Each stage should stop when context is cancelled.
//
// TODO: Implement stages that all respect the same context:
// - Stage 1: Generate numbers 1, 2, 3, ...
// - Stage 2: Square each number
// - Stage 3: Collect results
// All stages should stop promptly when ctx is cancelled.
//
// QUESTION: If stage 2 is slow, do stages 1 and 3 still respond to cancellation?
func ProcessPipeline(ctx context.Context) []int {
	// YOUR CODE HERE
	return nil
}

// =============================================================================
// PART 4: Concurrent Operations with Cancellation
// =============================================================================

// FirstSuccess runs multiple fetchers concurrently, returns first success.
// Cancels remaining fetchers once one succeeds.
//
// TODO: Implement to:
// 1. Create a cancellable context
// 2. Run all fetchers concurrently
// 3. Return first successful result (non-empty string)
// 4. Cancel context to stop other fetchers
// 5. Return "" if all fail or parent ctx cancelled
func FirstSuccess(ctx context.Context, fetchers []func(context.Context) string) string {
	// YOUR CODE HERE
	return ""
}

// AllSuccess runs fetchers concurrently, waits for all to complete.
// Cancels all if any fails or context is cancelled.
//
// TODO: Implement to:
// 1. Run all fetchers concurrently
// 2. If any returns "", cancel others and return nil
// 3. If ctx cancelled, return nil
// 4. Otherwise return all results
func AllSuccess(ctx context.Context, fetchers []func(context.Context) string) []string {
	// YOUR CODE HERE
	return nil
}

// =============================================================================
// PART 5: Context Values (Use Sparingly!)
// =============================================================================

// RequestID key for context values
type contextKey string

const RequestIDKey contextKey = "requestID"

// WithRequestID adds a request ID to the context.
func WithRequestID(ctx context.Context, id string) context.Context {
	// YOUR CODE HERE
	return ctx
}

// GetRequestID retrieves request ID from context.
// Returns "" if not present.
func GetRequestID(ctx context.Context) string {
	// YOUR CODE HERE
	return ""
}

// ProcessRequest simulates processing that logs with request ID.
//
// TODO: Implement to:
// 1. Extract request ID from context
// 2. Simulate work (respect cancellation!)
// 3. Return formatted result including request ID
//
// QUESTION: Why are context values controversial? When should you use them?
func ProcessRequest(ctx context.Context, data string) string {
	// YOUR CODE HERE
	return ""
}

// =============================================================================
// CHALLENGE: Implement graceful shutdown coordinator
// =============================================================================

// ShutdownCoordinator manages graceful shutdown of multiple services.
//
// TODO: Implement to:
// - Register services that need shutdown
// - Shutdown() cancels all services
// - Wait for all services to finish (with timeout)
// - Return errors from any service that failed to shutdown cleanly
type ShutdownCoordinator struct {
	// YOUR FIELDS HERE
}

// NewShutdownCoordinator creates a coordinator.
// func NewShutdownCoordinator() *ShutdownCoordinator {
// 	// YOUR CODE HERE
// 	return nil
// }

// Register adds a service. The service function should run until ctx is cancelled.
// func (sc *ShutdownCoordinator) Register(name string, service func(ctx context.Context) error) {
// 	// YOUR CODE HERE
// }

// Shutdown initiates shutdown and waits for completion.
// func (sc *ShutdownCoordinator) Shutdown(timeout time.Duration) map[string]error {
// 	// YOUR CODE HERE
// 	return nil
// }

// Ensure imports are used
var _ = context.Background
var _ = time.Second
