package concurrency

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// =============================================================================
// Tests for Exercise 6.1: Real-World Concurrency Patterns
// Run with: go test -v -run Test20
// =============================================================================

// --- PART 1: HTTP Handlers ---

func TestSlowHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "/?duration=50ms", nil)
	w := httptest.NewRecorder()

	done := make(chan struct{})
	go func() {
		SlowHandler(w, req)
		close(done)
	}()

	select {
	case <-done:
		if w.Code != 0 && w.Code != http.StatusOK {
			t.Logf("SlowHandler returned status %d", w.Code)
		}
	case <-time.After(5 * time.Second):
		t.Error("SlowHandler did not complete within 5 seconds")
	}
}

func TestSlowHandler_Cancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	req := httptest.NewRequest("GET", "/?duration=5s", nil).WithContext(ctx)
	w := httptest.NewRecorder()

	done := make(chan struct{})
	go func() {
		SlowHandler(w, req)
		close(done)
	}()

	// Cancel immediately
	cancel()

	select {
	case <-done:
		// Handler returned after cancellation - good
	case <-time.After(time.Second):
		t.Error("SlowHandler did not respect context cancellation")
	}
}

func TestTimeoutMiddleware(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}

	wrapped := TimeoutMiddleware(time.Second, handler)
	if wrapped == nil {
		t.Fatal("TimeoutMiddleware returned nil")
	}

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	wrapped(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d; want 200", w.Code)
	}
}

func TestTimeoutMiddleware_Slow(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(5 * time.Second):
			w.WriteHeader(http.StatusOK)
		case <-r.Context().Done():
			return
		}
	}

	wrapped := TimeoutMiddleware(50*time.Millisecond, handler)
	if wrapped == nil {
		t.Fatal("TimeoutMiddleware returned nil")
	}

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	done := make(chan struct{})
	go func() {
		wrapped(w, req)
		close(done)
	}()

	select {
	case <-done:
		// Middleware returned after timeout
		if w.Code != http.StatusServiceUnavailable && w.Code != http.StatusGatewayTimeout {
			t.Logf("TimeoutMiddleware returned status %d on timeout", w.Code)
		}
	case <-time.After(2 * time.Second):
		t.Error("TimeoutMiddleware did not enforce timeout")
	}
}

// --- PART 2: DBPool ---

func TestDBPool_Basic(t *testing.T) {
	config := DBPoolConfig{
		MinConns:        1,
		MaxConns:        3,
		IdleTimeout:     time.Minute,
		MaxLifetime:     time.Hour,
		HealthCheckFreq: time.Minute,
	}

	nextID := 0
	factory := func() (*DBConn, error) {
		nextID++
		return &DBConn{ID: nextID, CreatedAt: time.Now()}, nil
	}

	pool, err := NewDBPool(config, factory)
	if err != nil {
		t.Fatalf("NewDBPool error: %v", err)
	}
	if pool == nil {
		t.Fatal("NewDBPool returned nil")
	}
	defer pool.Close()

	ctx := context.Background()

	conn, err := pool.Acquire(ctx)
	if err != nil {
		t.Fatalf("Acquire error: %v", err)
	}
	if conn == nil {
		t.Fatal("Acquire returned nil connection")
	}

	pool.Release(conn)
}

func TestDBPool_Stats(t *testing.T) {
	config := DBPoolConfig{
		MinConns: 1,
		MaxConns: 5,
	}

	factory := func() (*DBConn, error) {
		return &DBConn{CreatedAt: time.Now()}, nil
	}

	pool, err := NewDBPool(config, factory)
	if err != nil {
		t.Fatalf("NewDBPool error: %v", err)
	}
	if pool == nil {
		t.Fatal("NewDBPool returned nil")
	}
	defer pool.Close()

	stats := pool.Stats()
	_ = stats // At minimum, Stats() should not panic
}

// --- PART 3: Graceful Shutdown ---

func TestGracefulServer(t *testing.T) {
	server := NewGracefulServer()
	if server == nil {
		t.Fatal("NewGracefulServer returned nil")
	}

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error)
	go func() {
		done <- server.Start(ctx)
	}()

	// Give it time to start
	time.Sleep(50 * time.Millisecond)

	// Trigger shutdown
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Second)
	defer shutdownCancel()

	err := server.Shutdown(shutdownCtx)
	if err != nil {
		t.Logf("Shutdown error: %v (may be expected)", err)
	}
}

func TestShutdownCoordinator2(t *testing.T) {
	sc := NewShutdownCoordinator2()
	if sc == nil {
		t.Fatal("NewShutdownCoordinator2 returned nil")
	}

	// Test with empty coordinator
	errs := sc.Shutdown(time.Second)
	if errs == nil {
		// nil map is fine for no services
	}
}

// --- PART 4: ParallelAPIHandler ---

func TestParallelAPIHandler(t *testing.T) {
	handler := ParallelAPIHandler([]string{"svc1", "svc2"}, http.DefaultClient)
	if handler == nil {
		t.Fatal("ParallelAPIHandler returned nil")
	}
	// Verify it's a valid handler function (doesn't panic on creation)
}

// --- Conceptual tests ---

func TestRealWorldConcept(t *testing.T) {
	t.Log("CONCEPT: Real-world Go services combine multiple concurrency patterns")
	t.Log("         HTTP handlers use request context for cancellation")
	t.Log("         Connection pools manage shared resources")
	t.Log("         Graceful shutdown ensures clean resource cleanup")
}
