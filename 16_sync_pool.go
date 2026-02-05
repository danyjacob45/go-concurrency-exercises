package concurrency

import (
	"bytes"
	"sync"
)

// =============================================================================
// EXERCISE 5.1: sync.Pool for Object Reuse
// =============================================================================
//
// sync.Pool provides a cache of allocated objects that can be reused to reduce
// GC pressure. Objects may be removed from the pool at any time without notice.
//
// KEY CONCEPTS:
// - Pool.Get() returns an object (or creates new one via New func)
// - Pool.Put() returns an object to the pool for reuse
// - Objects in pool may be garbage collected at any time
// - Pools are safe for concurrent use
// - Best for temporary objects that are expensive to allocate
// - NOT for connection pools or resource management (use channels instead)
//
// =============================================================================

// =============================================================================
// PART 1: Basic Pool Usage
// =============================================================================

// BufferPool manages reusable byte buffers.
//
// TODO: Implement a pool that:
// 1. Creates *bytes.Buffer when pool is empty
// 2. Provides Get() that returns a ready-to-use buffer
// 3. Provides Put() that resets and returns buffer to pool
//
// QUESTION: Why do we reset the buffer before putting it back?
type BufferPool struct {
	pool sync.Pool
}

func NewBufferPool() *BufferPool {
	// YOUR CODE HERE
	return nil
}

func (bp *BufferPool) Get() *bytes.Buffer {
	// YOUR CODE HERE
	return nil
}

func (bp *BufferPool) Put(buf *bytes.Buffer) {
	// YOUR CODE HERE
}

// =============================================================================
// PART 2: Pool with Size Limits
// =============================================================================

// SizedBuffer is a buffer with a known capacity.
type SizedBuffer struct {
	Data []byte
	Cap  int
}

// SizedBufferPool manages buffers of a specific size.
// This prevents small buffers from being returned to a large-buffer pool.
//
// TODO: Implement to:
// 1. Create buffers of exactly 'size' capacity
// 2. Only accept buffers back if they match the expected size
// 3. Discard buffers that have grown too large
//
// QUESTION: What happens if you Put() a buffer that grew beyond the original size?
type SizedBufferPool struct {
	// YOUR FIELDS HERE
}

func NewSizedBufferPool(size int) *SizedBufferPool {
	// YOUR CODE HERE
	return nil
}

func (sp *SizedBufferPool) Get() []byte {
	// YOUR CODE HERE
	return nil
}

func (sp *SizedBufferPool) Put(buf []byte) {
	// YOUR CODE HERE
}

// =============================================================================
// PART 3: Pool Statistics
// =============================================================================

// TrackedPool tracks pool usage statistics.
// Useful for understanding pool effectiveness.
//
// TODO: Implement a pool that tracks:
// - Total Get() calls
// - Cache hits (reused from pool)
// - Cache misses (created new)
// - Total Put() calls
//
// NOTE: Stats must be thread-safe!
type TrackedPool struct {
	// YOUR FIELDS HERE
}

type SyncPoolStats struct {
	Gets   int64
	Hits   int64
	Misses int64
	Puts   int64
}

func NewTrackedPool(newFunc func() any) *TrackedPool {
	// YOUR CODE HERE
	return nil
}

func (tp *TrackedPool) Get() any {
	// YOUR CODE HERE
	return nil
}

func (tp *TrackedPool) Put(x any) {
	// YOUR CODE HERE
}

func (tp *TrackedPool) Stats() SyncPoolStats {
	// YOUR CODE HERE
	return SyncPoolStats{}
}

// =============================================================================
// PART 4: Common Pool Patterns
// =============================================================================

// JSONEncoderPool pools JSON encoders for reuse.
// Encoders are expensive to create due to reflection caching.
//
// TODO: Implement pool for json.Encoder
// HINT: Each Get() needs a fresh bytes.Buffer, but encoder can be reused
// QUESTION: Can you actually pool json.Encoder? What's the challenge?

// SlicePool manages reusable slices of a specific type.
//
// TODO: Implement a generic-style pool for []int slices
// 1. Get() returns a slice with len=0 but existing capacity
// 2. Put() should clear the slice (set len to 0, keep capacity)
type IntSlicePool struct {
	// YOUR FIELDS HERE
}

func NewIntSlicePool(defaultCap int) *IntSlicePool {
	// YOUR CODE HERE
	return nil
}

func (p *IntSlicePool) Get() []int {
	// YOUR CODE HERE
	return nil
}

func (p *IntSlicePool) Put(s []int) {
	// YOUR CODE HERE
}

// =============================================================================
// PART 5: When NOT to Use Pool
// =============================================================================

// AntiPattern1: Don't pool small objects
// The overhead of pool operations exceeds allocation cost for small objects.
// Rule of thumb: only pool objects > 1KB or with expensive initialization.

// AntiPattern2: Don't pool objects with state dependencies
// If object state depends on external factors, pooling can cause bugs.

// AntiPattern3: Don't use Pool for resource management
// Connections, file handles, etc. need deterministic lifecycle management.
// Use a channel-based pool or dedicated pool library instead.

// ConnectionPool demonstrates the WRONG way to pool connections.
// QUESTION: Why is sync.Pool wrong for database connections?
//
// type BadConnectionPool struct {
// 	pool sync.Pool // DON'T DO THIS
// }

// GoodConnectionPool shows the correct pattern using channels.
//
// TODO: Implement a proper connection pool using channels
// 1. Fixed maximum number of connections
// 2. Get() blocks if no connection available
// 3. Put() returns connection to pool
// 4. Close() closes all connections
type PooledConn struct {
	ID     int
	closed bool
}

func (c *PooledConn) Close() error {
	c.closed = true
	return nil
}

type GoodConnectionPool struct {
	// YOUR FIELDS HERE
}

func NewGoodConnectionPool(maxConns int, factory func(id int) *PooledConn) *GoodConnectionPool {
	// YOUR CODE HERE
	return nil
}

func (p *GoodConnectionPool) Get() *PooledConn {
	// YOUR CODE HERE
	return nil
}

func (p *GoodConnectionPool) Put(c *PooledConn) {
	// YOUR CODE HERE
}

func (p *GoodConnectionPool) Close() {
	// YOUR CODE HERE
}

// =============================================================================
// CHALLENGE: Implement a multi-size buffer pool
// =============================================================================

// MultiSizePool manages buffers of different size classes.
// Like how memory allocators work (e.g., TCMalloc).
//
// TODO: Implement to:
// 1. Define size classes: 1KB, 4KB, 16KB, 64KB
// 2. Get(size) returns buffer from smallest class that fits
// 3. Put() returns buffer to appropriate class based on capacity
// 4. Track stats per size class
//
// QUESTION: Why use size classes instead of exact sizes?
type MultiSizePool struct {
	// YOUR FIELDS HERE
}

// func NewMultiSizePool() *MultiSizePool {
// 	// YOUR CODE HERE
// 	return nil
// }

// func (p *MultiSizePool) Get(size int) []byte {
// 	// YOUR CODE HERE
// 	return nil
// }

// func (p *MultiSizePool) Put(buf []byte) {
// 	// YOUR CODE HERE
// }

// Ensure imports are used
var _ = bytes.Buffer{}
var _ = sync.Pool{}
