package concurrency

import (
	"sync"
	"time"
)

// =============================================================================
// EXERCISE 2.2: sync.Mutex vs sync.RWMutex
// =============================================================================
//
// Mutexes provide mutual exclusion for protecting shared data.
//
// KEY CONCEPTS:
// - Mutex: only one goroutine can hold the lock at a time
// - RWMutex: multiple readers OR one writer (but not both)
// - Always unlock in the same goroutine that locked
// - Use defer mu.Unlock() to ensure unlock even on panic
// - Copying a mutex after first use is forbidden (pass by pointer!)
//
// =============================================================================

// =============================================================================
// PART 1: Basic Mutex - Thread-Safe Counter
// =============================================================================

// Counter is a thread-safe counter.
//
// TODO: Implement all methods to be goroutine-safe
type Counter struct {
	// YOUR FIELDS HERE
}

// NewCounter creates a new counter starting at 0.
func NewCounter() *Counter {
	// YOUR CODE HERE
	return nil
}

// Inc increments the counter by 1.
func (c *Counter) Inc() {
	// YOUR CODE HERE
}

// Dec decrements the counter by 1.
func (c *Counter) Dec() {
	// YOUR CODE HERE
}

// Value returns the current count.
func (c *Counter) Value() int {
	// YOUR CODE HERE
	return 0
}

// =============================================================================
// PART 2: RWMutex - Thread-Safe Cache
// =============================================================================

// Cache is a thread-safe string cache optimized for read-heavy workloads.
//
// TODO: Implement using RWMutex
// - Reads should not block other reads
// - Writes should have exclusive access
//
// QUESTION: When would you use Mutex instead of RWMutex?
type Cache struct {
	// YOUR FIELDS HERE
}

// NewCache creates a new empty cache.
func NewCache() *Cache {
	// YOUR CODE HERE
	return nil
}

// Get retrieves a value from the cache.
// Returns (value, true) if found, ("", false) if not.
func (c *Cache) Get(key string) (string, bool) {
	// YOUR CODE HERE
	return "", false
}

// Set stores a value in the cache.
func (c *Cache) Set(key, value string) {
	// YOUR CODE HERE
}

// Delete removes a key from the cache.
func (c *Cache) Delete(key string) {
	// YOUR CODE HERE
}

// Len returns the number of items in the cache.
func (c *Cache) Len() int {
	// YOUR CODE HERE
	return 0
}

// =============================================================================
// PART 3: Lock Ordering to Prevent Deadlocks
// =============================================================================

// Account represents a bank account.
type Account struct {
	ID      int
	balance int
	mu      sync.Mutex
}

// NewAccount creates an account with initial balance.
func NewAccount(id, balance int) *Account {
	return &Account{ID: id, balance: balance}
}

// Balance returns current balance.
func (a *Account) Balance() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.balance
}

// Transfer moves amount from account 'from' to account 'to'.
//
// TODO: Implement this WITHOUT deadlocks!
//
// WRONG approach (causes deadlock):
//   from.mu.Lock()
//   to.mu.Lock()    // If another goroutine does Transfer(to, from, x), deadlock!
//
// QUESTION: How do you prevent deadlock when locking multiple mutexes?
// HINT: Always acquire locks in a consistent order (e.g., by account ID)
func Transfer(from, to *Account, amount int) bool {
	// YOUR CODE HERE
	return false
}

// =============================================================================
// PART 4: Mutex Patterns
// =============================================================================

// SafeSlice is a thread-safe slice.
//
// TODO: Implement append, get, and len operations
type SafeSlice struct {
	// YOUR FIELDS HERE
}

// NewSafeSlice creates an empty thread-safe slice.
func NewSafeSlice() *SafeSlice {
	// YOUR CODE HERE
	return nil
}

// Append adds a value to the slice.
func (s *SafeSlice) Append(val int) {
	// YOUR CODE HERE
}

// Get returns value at index. Panics if out of bounds.
func (s *SafeSlice) Get(index int) int {
	// YOUR CODE HERE
	return 0
}

// Len returns the length of the slice.
func (s *SafeSlice) Len() int {
	// YOUR CODE HERE
	return 0
}

// =============================================================================
// CHALLENGE: Implement a Read-Write cache with expiration
// =============================================================================

// ExpiringCache stores values that expire after a duration.
//
// TODO: Implement a cache where:
// - Each entry has a timestamp
// - Get returns ("", false) for expired entries
// - Set stores value with current time
// - CleanExpired removes all expired entries
//
// HINT: You'll need time.Time for timestamps and time.Duration for expiry
type ExpiringCache struct {
	// YOUR FIELDS HERE
}

// NewExpiringCache creates a cache with given TTL for entries.
func NewExpiringCache(ttl time.Duration) *ExpiringCache {
    return nil // TODO: implement
}

func (c *ExpiringCache) Set(key int, value string) {}

func (c *ExpiringCache) Get(key int) (string, bool) { return "", false }

func (c *ExpiringCache) CleanExpired() {}

// Ensure sync import is used
var _ = sync.Mutex{}
