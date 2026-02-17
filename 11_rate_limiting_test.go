package concurrency

import (
	"context"
	"testing"
	"time"
)

// =============================================================================
// Tests for Exercise 3.4: Rate Limiting
// Run with: go test -v -run Test11
// =============================================================================

// --- PART 1: TickerLimiter ---

func TestTickerLimiter(t *testing.T) {
	limiter := NewTickerLimiter(50 * time.Millisecond)
	if limiter == nil {
		t.Fatal("NewTickerLimiter returned nil")
	}
	defer limiter.Stop()

	start := time.Now()

	// Should be rate limited
	limiter.Wait()
	limiter.Wait()
	limiter.Wait()

	elapsed := time.Since(start)

	// 3 waits at 50ms interval should take at least ~100ms (first may be immediate)
	if elapsed < 80*time.Millisecond {
		t.Errorf("3 waits took %v; expected at least ~100ms", elapsed)
	}
}

func TestTickerLimiter_Stop(t *testing.T) {
	limiter := NewTickerLimiter(10 * time.Millisecond)
	if limiter == nil {
		t.Fatal("NewTickerLimiter returned nil")
	}

	limiter.Stop()
	// Should not panic on double stop or use after stop
}

// --- PART 2: TokenBucket ---

func TestTokenBucket_TryTake(t *testing.T) {
	tb := NewTokenBucket(5, 10.0) // capacity 5, 10 tokens/sec
	if tb == nil {
		t.Fatal("NewTokenBucket returned nil")
	}

	// Bucket starts full (5 tokens)
	if !tb.TryTake(3) {
		t.Error("TryTake(3) failed; bucket should have 5 tokens")
	}

	if !tb.TryTake(2) {
		t.Error("TryTake(2) failed; bucket should have 2 tokens left")
	}

	// Bucket should be empty now
	if tb.TryTake(1) {
		t.Error("TryTake(1) succeeded; bucket should be empty")
	}
}

func TestTokenBucket_Available(t *testing.T) {
	tb := NewTokenBucket(10, 1.0)
	if tb == nil {
		t.Fatal("NewTokenBucket returned nil")
	}

	initial := tb.Available()
	if initial != 10 {
		t.Errorf("initial Available() = %d; want 10", initial)
	}

	tb.TryTake(5)
	after := tb.Available()
	if after != 5 {
		t.Errorf("Available() after TryTake(5) = %d; want 5", after)
	}
}

func TestTokenBucket_Take_Blocks(t *testing.T) {
	tb := NewTokenBucket(1, 100.0) // 1 capacity, fast refill
	if tb == nil {
		t.Fatal("NewTokenBucket returned nil")
	}

	// Drain the bucket
	tb.TryTake(1)

	// Take should block until refill
	done := make(chan struct{})
	go func() {
		tb.Take(1)
		close(done)
	}()

	select {
	case <-done:
		// OK, Take completed after refill
	case <-time.After(1 * time.Second):
		t.Error("Take(1) blocked for over 1 second; expected refill")
	}
}

// --- PART 3: ContextLimiter ---

func TestContextLimiter_CancelledContext(t *testing.T) {
	cl := NewContextLimiter(1.0) // 1 op/sec
	if cl == nil {
		t.Fatal("NewContextLimiter returned nil")
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := cl.WaitN(ctx, 1)
	if err == nil {
		t.Error("WaitN with cancelled context should return error")
	}
}

func TestContextLimiter_Success(t *testing.T) {
	cl := NewContextLimiter(1000.0) // Very fast rate
	if cl == nil {
		t.Fatal("NewContextLimiter returned nil")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := cl.WaitN(ctx, 1)
	if err != nil {
		t.Errorf("WaitN returned error: %v", err)
	}
}

// --- PART 4: PerKeyLimiter ---

func TestPerKeyLimiter(t *testing.T) {
	pkl := NewPerKeyLimiter(2, 1.0) // 2 capacity per key
	if pkl == nil {
		t.Fatal("NewPerKeyLimiter returned nil")
	}

	// Different keys should have independent limits
	if !pkl.Allow("user1") {
		t.Error("Allow(user1) first call returned false")
	}
	if !pkl.Allow("user1") {
		t.Error("Allow(user1) second call returned false")
	}

	// user1 should be exhausted
	if pkl.Allow("user1") {
		t.Error("Allow(user1) third call returned true; bucket should be empty")
	}

	// user2 should still have tokens
	if !pkl.Allow("user2") {
		t.Error("Allow(user2) first call returned false; should be independent")
	}
}

// --- PART 5: LeakyBucket ---

func TestLeakyBucket_Add(t *testing.T) {
	lb := NewLeakyBucket(3, 10.0)
	if lb == nil {
		t.Fatal("NewLeakyBucket returned nil")
	}

	if !lb.Add(1) {
		t.Error("Add(1) returned false; bucket should have space")
	}
	if !lb.Add(2) {
		t.Error("Add(2) returned false; bucket should have space")
	}
	if !lb.Add(3) {
		t.Error("Add(3) returned false; bucket should have space")
	}

	// Bucket full
	if lb.Add(4) {
		t.Error("Add(4) returned true; bucket should be full")
	}
}

func TestLeakyBucket_StartStop(t *testing.T) {
	lb := NewLeakyBucket(5, 100.0) // Fast drain
	if lb == nil {
		t.Fatal("NewLeakyBucket returned nil")
	}

	var processed []int
	done := make(chan struct{})

	lb.Add(1)
	lb.Add(2)
	lb.Add(3)

	go func() {
		lb.Start(func(item int) {
			processed = append(processed, item)
		})
		close(done)
	}()

	time.Sleep(200 * time.Millisecond)
	lb.Stop()

	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Error("LeakyBucket.Start did not return after Stop")
	}

	if len(processed) == 0 {
		t.Error("no items were processed")
	}
}

// --- Conceptual tests ---

func TestRateLimitingConcept(t *testing.T) {
	t.Log("CONCEPT: Rate limiting controls how frequently operations can occur")
	t.Log("         Ticker: steady rate, one op per interval")
	t.Log("         Token bucket: allows bursts up to capacity")
	t.Log("         Leaky bucket: smooths out bursts completely")
}
