package concurrency

// =============================================================================
// EXERCISE 4.1: Race Condition Detection
// =============================================================================
//
// A race condition occurs when multiple goroutines access shared data
// concurrently and at least one access is a write.
//
// KEY CONCEPTS:
// - Race detector: go run -race, go test -race
// - Races are undefined behavior - program may crash, corrupt data, or "work"
// - "Working" code with races is the most dangerous - hides bugs
// - Fix options: mutex, atomic, channels, or restructure to avoid sharing
//
// =============================================================================

// =============================================================================
// PART 1: Identify the Race
// =============================================================================

// EXERCISE: Each function below has a race condition.
// 1. Run with: go test -race -run TestRace
// 2. Identify WHY it's a race
// 3. Write the fixed version in the corresponding Fix function

// RaceCounter has a race condition.
// QUESTION: What's the race? What values might it return?
func RaceCounter() int {
	counter := 0
	for i := 0; i < 1000; i++ {
		go func() {
			counter++ // RACE: concurrent read-modify-write
		}()
	}
	// Wait a bit for goroutines (bad practice, just for demo)
	// time.Sleep(100 * time.Millisecond)
	return counter
}

// RaceCounterFix should return exactly 1000.
//
// TODO: Fix using ONE of these approaches:
// - sync.Mutex
// - sync/atomic
// - Channel to serialize updates
func RaceCounterFix() int {
	// YOUR CODE HERE
	return 0
}

// RaceSlice has a race condition.
// QUESTION: Why is appending to a slice not safe?
func RaceSlice() []int {
	var results []int
	for i := 0; i < 100; i++ {
		go func(n int) {
			results = append(results, n) // RACE: slice header + backing array
		}(i)
	}
	// time.Sleep(100 * time.Millisecond)
	return results
}

// RaceSliceFix should return a slice with 100 elements.
//
// TODO: Fix the race (hint: mutex around append, or use channel)
func RaceSliceFix() []int {
	// YOUR CODE HERE
	return nil
}

// RaceMap has a race condition.
// QUESTION: Why do maps have explicit race detection in Go?
func RaceMap() map[int]int {
	m := make(map[int]int)
	for i := 0; i < 100; i++ {
		go func(n int) {
			m[n] = n * n // RACE: concurrent map write
		}(i)
	}
	// time.Sleep(100 * time.Millisecond)
	return m
}

// RaceMapFix should return a map with 100 entries.
//
// TODO: Fix using ONE of:
// - sync.Mutex
// - sync.Map (built-in concurrent map)
func RaceMapFix() map[int]int {
	// YOUR CODE HERE
	return nil
}

// =============================================================================
// PART 2: Subtle Races
// =============================================================================

// RaceCheckThenAct demonstrates check-then-act race.
// Even though we check first, another goroutine might change value between
// check and act.
type UnsafeCounter struct {
	value int
}

func (c *UnsafeCounter) IncrementIfBelow(max int) bool {
	if c.value < max { // CHECK
		c.value++ // ACT - race! value might have changed
		return true
	}
	return false
}

// SafeCounter fixes the check-then-act race.
//
// TODO: Implement thread-safe version
// HINT: The check AND act must be atomic (under same lock)
type SafeCounter struct {
	// YOUR FIELDS HERE
}

func NewSafeCounter() *SafeCounter {
	// YOUR CODE HERE
	return nil
}

func (c *SafeCounter) IncrementIfBelow(max int) bool {
	// YOUR CODE HERE
	return false
}

func (c *SafeCounter) Value() int {
	// YOUR CODE HERE
	return 0
}

// =============================================================================
// PART 3: Race in Loop Variable
// =============================================================================

// RaceLoopVar demonstrates the classic loop variable capture bug.
//
// IMPORTANT GO VERSION NOTE:
// - Go 1.21 and earlier: Loop variable 'i' is shared across iterations.
//   All goroutines capture the SAME variable, usually seeing the final value (5).
// - Go 1.22+: Each iteration creates a NEW loop variable.
//   This bug is now FIXED at the language level for for-range loops!
//
// This exercise still demonstrates the concept, but if you're using Go 1.22+,
// you may not see the bug. The fix patterns shown below are still useful for
// understanding and for code that must work with older Go versions.
//
// QUESTION: What values get printed? Why?
func RaceLoopVar() []int {
	var results []int
	for i := 0; i < 5; i++ {
		go func() {
			results = append(results, i) // Go <1.22: Captures 'i' by reference!
		}()
	}
	// time.Sleep(100 * time.Millisecond)
	return results
}

// RaceLoopVarFix should return [0, 1, 2, 3, 4] (in some order).
//
// TODO: Fix the loop variable capture bug
// HINT: Two ways - parameter or local copy
func RaceLoopVarFix() []int {
	// YOUR CODE HERE
	return nil
}

// =============================================================================
// PART 4: Race with Struct Fields
// =============================================================================

// Stats tracks statistics with a race condition.
type Stats struct {
	Count int
	Sum   int
	Max   int
}

// RaceStructFields has multiple race conditions.
func RaceStructFields() Stats {
	stats := Stats{}
	for i := 1; i <= 100; i++ {
		go func(n int) {
			stats.Count++
			stats.Sum += n
			if n > stats.Max {
				stats.Max = n
			}
		}(i)
	}
	// time.Sleep(100 * time.Millisecond)
	return stats
}

// SafeStats is a thread-safe version of Stats.
//
// TODO: Implement with proper synchronization
// QUESTION: Should each field have its own lock? Why or why not?
type SafeStats struct {
	// YOUR FIELDS HERE
}

func NewSafeStats() *SafeStats {
	// YOUR CODE HERE
	return nil
}

func (s *SafeStats) Record(n int) {
	// YOUR CODE HERE
}

func (s *SafeStats) Snapshot() Stats {
	// YOUR CODE HERE
	return Stats{}
}

// =============================================================================
// CHALLENGE: Find the race in production-like code
// =============================================================================

// Cache implements a cache with race conditions.
// This looks like real code you might write!
//
// TODO: Find ALL race conditions and fix them
type RacyCache struct {
	data    map[string]string
	hits    int
	misses  int
}

func NewRacyCache() *RacyCache {
	return &RacyCache{data: make(map[string]string)}
}

func (c *RacyCache) Get(key string) (string, bool) {
	val, ok := c.data[key] // RACE 1
	if ok {
		c.hits++ // RACE 2
	} else {
		c.misses++ // RACE 3
	}
	return val, ok
}

func (c *RacyCache) Set(key, value string) {
	c.data[key] = value // RACE 4
}

func (c *RacyCache) Stats() (hits, misses int) {
	return c.hits, c.misses // RACE 5, 6
}

// SafeCache is the fixed version.
//
// TODO: Implement without any race conditions
type SafeCache struct {
	// YOUR FIELDS HERE
}

// func NewSafeCache() *SafeCache {
// 	// YOUR CODE HERE
// 	return nil
// }
