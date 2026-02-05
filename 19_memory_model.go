package concurrency

import (
	"sync"
	"sync/atomic"
)

// =============================================================================
// EXERCISE 5.4: Go Memory Model and Happens-Before
// =============================================================================
//
// The Go memory model specifies when writes in one goroutine are visible
// to reads in another. Understanding this prevents subtle bugs.
//
// REFERENCE: https://go.dev/ref/mem (updated in Go 1.19 with clarifications)
//
// KEY CONCEPTS:
// - "Happens-before" is a partial ordering of memory operations
// - If A happens-before B, A's effects are visible to B
// - Without synchronization, there's NO guarantee about visibility
// - Synchronization primitives establish happens-before relationships
//
// PRACTICAL TIP: Always run tests with the race detector!
//   go test -race ./...
//   go run -race main.go
// The race detector finds data races at runtime. It won't catch everything
// (only races that actually occur during execution), but it's invaluable.
//
// HAPPENS-BEFORE RULES:
// 1. Within a goroutine: statements happen in program order
// 2. Channel send happens-before corresponding receive completes
// 3. Channel close happens-before receive that returns zero value
// 4. For unbuffered channels: receive happens-before send completes
//    (More precisely: the k-th receive on a channel with capacity C
//    happens-before the (k+C)-th send completes)
// 5. Lock() happens-before any subsequent Unlock() on the same mutex
//    Unlock() happens-before any subsequent Lock() on the same mutex
// 6. sync.Once: the single call to f() in once.Do(f) happens-before
//    any call to once.Do() returns
// 7. Atomic operations on the same memory location have a total order
//    that all goroutines agree on (sequential consistency)
// 8. Goroutine creation: the go statement happens-before the new
//    goroutine begins execution
// 9. Package init: all init() functions happen-before main() begins
//
// =============================================================================

// =============================================================================
// PART 1: Visibility Without Synchronization
// =============================================================================

// UnsafeVisibility demonstrates that without sync, changes may not be visible.
//
// QUESTION: What might secondGoroutineSaw return? (0? 42? something else?)
// ANSWER: ANY of these! Without synchronization, there's no guarantee.
func UnsafeVisibility() int {
	var value int
	var seen int

	go func() {
		value = 42
	}()

	go func() {
		seen = value // May see 0 or 42 - undefined!
	}()

	// Don't do this - just for demonstration
	// time.Sleep(time.Millisecond)

	return seen // This is also racy!
}

// ChannelVisibility demonstrates happens-before with channels.
//
// TODO: Fix UnsafeVisibility using a channel to establish happens-before.
// The second goroutine should ALWAYS see value = 42.
func ChannelVisibility() int {
	// YOUR CODE HERE
	return 0
}

// MutexVisibility demonstrates happens-before with mutex.
//
// TODO: Fix UnsafeVisibility using mutex.
func MutexVisibility() int {
	// YOUR CODE HERE
	return 0
}

// =============================================================================
// PART 2: The "Happens-Before" Relationship
// =============================================================================

// Understanding which operations establish happens-before.

// ChannelHappensBefore demonstrates the channel happens-before rules.
//
// Rule: A send on a channel happens-before the corresponding receive completes.
//
// EXPLANATION: The happens-before chain is:
//   a = 1              (sequenced-before in goroutine)
//   b = 2              (sequenced-before in goroutine)
//   ch <- struct{}{}   (send happens-before receive completes)
//   <-ch completes     (in main goroutine)
//   return a, b        (sequenced-after in main goroutine)
//
// Because happens-before is transitive: a=1 and b=2 happen-before return a, b
func ChannelHappensBefore() (int, int) {
	var a, b int
	ch := make(chan struct{})

	go func() {
		a = 1
		b = 2
		ch <- struct{}{} // Send happens-before...
	}()

	<-ch        // ...receive completes
	return a, b // Guaranteed to be 1, 2
}

// UnbufferedChannelOrder demonstrates receive-before-send-completes rule.
//
// Rule: For unbuffered channels, receive happens-before send COMPLETES.
//
// EXPLANATION: This is the "reverse" synchronization that unbuffered channels
// provide. The main goroutine's send BLOCKS until the spawned goroutine's
// receive is ready. Therefore:
//
//   value = 42          (in goroutine)
//   <-ch                (in goroutine, receive starts)
//   ch <- struct{}{}    (in main, send completes AFTER receive)
//   return value        (in main)
//
// The receive happens-before the send completes, and value=42 is sequenced
// before the receive, so value=42 happens-before return value.
func UnbufferedChannelOrder() int {
	var value int
	ch := make(chan struct{})

	go func() {
		value = 42
		<-ch // Receive happens first...
	}()

	ch <- struct{}{} // ...then send completes
	return value     // Guaranteed to be 42
}

// BufferedChannelOrder shows that buffered channels are different.
//
// QUESTION: Is this code correct? Why or why not?
//
// ANSWER: YES, this is correct! The rule "send happens-before receive completes"
// applies to ALL channels, buffered or unbuffered. The send of struct{}{} happens
// before the receive completes, so the write to value (which happens before the
// send) is guaranteed to be visible after the receive.
//
// The difference with buffered channels is rule #4: for unbuffered channels,
// receive also happens-before send COMPLETES (bidirectional sync). Buffered
// channels don't have this reverse guarantee until the buffer is full.
func BufferedChannelOrder() int {
	var value int
	ch := make(chan struct{}, 1)

	go func() {
		value = 42
		ch <- struct{}{} // Send happens-before receive completes
	}()

	<-ch
	return value // Guaranteed to be 42
}

// =============================================================================
// PART 3: Atomic Operations and Memory Ordering
// =============================================================================

// AtomicVisibility demonstrates atomic operations provide visibility.
//
// Atomic operations create a total order that all goroutines agree on.
//
// EXPLANATION: This is correct because:
// 1. atomic.StoreInt64(&value, 42) is sequenced-before atomic.StoreInt64(&ready, 1)
// 2. The Load of ready eventually observes the Store of 1
// 3. The Store of 1 is synchronized-before the Load that observes it
// 4. Therefore, Store(&value, 42) happens-before Load(&value)
//
// WARNING: Spinning (busy-waiting) wastes CPU. In real code, use channels,
// sync.Cond, or other blocking primitives. This is just for demonstration.
func AtomicVisibility() int64 {
	var value int64
	var ready int64

	go func() {
		atomic.StoreInt64(&value, 42)
		atomic.StoreInt64(&ready, 1)
	}()

	// Spin until ready (don't do this in real code!)
	for atomic.LoadInt64(&ready) == 0 {
		// busy wait
	}

	return atomic.LoadInt64(&value) // Guaranteed to be 42
}

// AtomicWrongUsage shows a common mistake.
//
// QUESTION: What's wrong with this code?
//
// NUANCED ANSWER (Go 1.19+ memory model clarification):
// This code is TECHNICALLY safe for visibility in Go 1.19+. The memory model
// was clarified: if atomic operation A is sequenced-before atomic operation B,
// and B is observed by atomic operation C, then A happens-before C. This means
// the non-atomic write to `value` (before the atomic store to `ready`) will be
// visible after the atomic load of `ready` observes the value 1.
//
// HOWEVER, this is still considered BAD PRACTICE because:
// 1. The race detector will flag it as a data race (non-atomic + atomic on different vars)
// 2. It's fragile - mixing atomic and non-atomic access patterns is error-prone
// 3. It relies on subtle memory model guarantees that are easy to get wrong
// 4. If you later access `value` from another goroutine without the ready check, it's racy
//
// BEST PRACTICE: Use atomic operations consistently for ALL shared variables,
// or use higher-level synchronization (channels, mutex).
func AtomicWrongUsage() int64 {
	var value int64
	var ready int64

	go func() {
		value = 42 // Non-atomic write!
		atomic.StoreInt64(&ready, 1)
	}()

	for atomic.LoadInt64(&ready) == 0 {
	}

	return value // Non-atomic read - technically visible, but bad practice
}

// AtomicFix fixes the above code.
//
// TODO: Make both accesses atomic
func AtomicFix() int64 {
	// YOUR CODE HERE
	return 0
}

// =============================================================================
// PART 4: sync.Once Guarantees
// =============================================================================

// OnceHappensBefore demonstrates sync.Once provides happens-before.
//
// Rule: The function passed to once.Do() happens-before any Do() returns.
//
// TODO: Implement singleton initialization using sync.Once
// Multiple goroutines should all see the fully initialized value.
type ExpensiveSingleton struct {
	data []int
}

var (
	singletonInstance *ExpensiveSingleton
	singletonOnce     sync.Once
)

func GetSingletonInstance() *ExpensiveSingleton {
	// YOUR CODE HERE
	return nil
}

// =============================================================================
// PART 5: Publication Safety
// =============================================================================

// SafePublication demonstrates how to safely publish an object.
// "Publication" means making a reference visible to other goroutines.
//
// WRONG: Just assigning the pointer (no happens-before)
// RIGHT: Use channel, mutex, or atomic to establish happens-before

type AppConfig struct {
	Servers []string
	Timeout int
}

// UnsafePublication demonstrates incorrect publication.
//
// WARNING: Other goroutines might see:
// - nil (the write hasn't propagated yet)
// - A valid pointer but with zero/partial field values
// - The Servers slice header without the backing array data
//
// This happens because the compiler and CPU can reorder memory operations.
// The pointer assignment might become visible before the field assignments!
var unsafeConfig *AppConfig

func UnsafePublish() {
	cfg := &AppConfig{
		Servers: []string{"a", "b", "c"},
		Timeout: 30,
	}
	unsafeConfig = cfg // WRONG: no synchronization
}

// SafePublicationAtomic uses atomic.Value for safe publication.
//
// TODO: Implement safe publication using atomic.Value
var safeAppConfig atomic.Value // Will hold *AppConfig

func SafePublish() {
	// YOUR CODE HERE
}

func GetAppConfig() *AppConfig {
	// YOUR CODE HERE
	return nil
}

// =============================================================================
// PART 6: Practical Examples
// =============================================================================

// LazyInitMap demonstrates safe lazy initialization of a map.
//
// TODO: Implement a map that initializes lazily in a thread-safe way
// Use sync.RWMutex with double-check locking pattern
//
// DOUBLE-CHECKED LOCKING PATTERN:
// 1. Acquire read lock, check if key exists
// 2. If exists, return value (fast path - multiple readers allowed)
// 3. If not, release read lock, acquire write lock
// 4. Check AGAIN if key exists (another goroutine may have created it)
// 5. If still not exists, create it
// 6. Release write lock, return value
//
// NOTE: In Go, unlike Java/C++, double-checked locking with RWMutex is safe
// because Unlock() happens-before subsequent Lock()/RLock() on the same mutex.
type LazyMap struct {
	// YOUR FIELDS HERE
	// Hint: you need sync.RWMutex and map[string]any
}

func NewLazyMap() *LazyMap {
	// YOUR CODE HERE
	return nil
}

func (m *LazyMap) GetOrCreate(key string, create func() any) any {
	// YOUR CODE HERE
	return nil
}

// =============================================================================
// CHALLENGE: Implement a lock-free stack
// =============================================================================

// AtomicStack implements a stack using atomic compare-and-swap.
// This is an advanced topic - understand happens-before first!
//
// TODO: Implement using atomic.Pointer[T] (Go 1.19+) or unsafe.Pointer with
// atomic.CompareAndSwapPointer
//
// Requirements:
// 1. Push adds to top of stack
// 2. Pop removes and returns top
// 3. No locks allowed!
//
// ALGORITHM (Treiber Stack):
// Push: 1. Read current head
//       2. Create new node pointing to current head
//       3. CAS head from old to new; if fails, retry from step 1
//
// Pop:  1. Read current head
//       2. If nil, stack is empty
//       3. CAS head from current to current.next; if fails, retry from step 1
//       4. Return the value from the old head
//
// IMPORTANT CONSIDERATIONS:
// - ABA Problem: In languages with manual memory management, a node could be
//   popped, freed, reallocated, and pushed again. CAS would succeed incorrectly.
//   Go's GC prevents this specific issue, but be aware of it for other languages.
// - Memory Ordering: atomic.Pointer operations provide the necessary happens-before
//   guarantees for the node's contents to be visible.
// - Spurious CAS failures: Always use a retry loop.
type AtomicStackNode struct {
	value any
	next  *AtomicStackNode
}

type AtomicStack struct {
	// YOUR FIELDS HERE
}

func NewAtomicStack() *AtomicStack {
	// YOUR CODE HERE
	return nil
}

func (s *AtomicStack) Push(value any) {
	// YOUR CODE HERE
}

func (s *AtomicStack) Pop() (any, bool) {
	// YOUR CODE HERE
	return nil, false
}

// Ensure imports are used
var _ = sync.Mutex{}
var _ = atomic.Value{}
