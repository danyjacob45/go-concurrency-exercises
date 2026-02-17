package concurrency

import (
	"time"
)

// =============================================================================
// EXERCISE 1.1: Channel Semantics
// =============================================================================
//
// This exercise explores the fundamental differences between buffered and
// unbuffered channels, and how channels behave when closed.
//
// KEY CONCEPTS:
// - Unbuffered channels block until both sender AND receiver are ready
// - Buffered channels block only when the buffer is full (send) or empty (receive)
// - Receiving from a closed channel returns the zero value immediately
// - Sending to a closed channel panics
// - You can detect closure with the "comma ok" idiom: val, ok := <-ch
//
// =============================================================================

// UnbufferedDemo demonstrates unbuffered channel behavior.
// An unbuffered channel blocks the sender until a receiver is ready.
//
// TODO: Implement this function to:
// 1. Create an unbuffered channel of int
// 2. Launch a goroutine that sends the value 42 to the channel
// 3. Sleep for 100ms in the main goroutine (simulating work)
// 4. Receive from the channel and return the value
//
// QUESTION: Where does the goroutine block? Before or after the sleep?
func UnbufferedDemo() int {
	// YOUR CODE HERE
	return 0
}

// BufferedDemo demonstrates buffered channel behavior.
// A buffered channel allows sends to complete without blocking (until full).
func BufferedDemo() int {
	// YOUR CODE HERE
	return 0
}

// BufferFullDemo demonstrates what happens when a buffer fills up.
//
// TODO: Implement this function to:
// 1. Create a buffered channel of int with capacity 2
// 2. Send values 1, 2 to the channel (fills the buffer)
// 3. Launch a goroutine that will receive one value after 50ms
// 4. Send value 3 (this should block until the goroutine receives!)
// 5. Return true if all sends completed successfully
//
// QUESTION: What would happen if you didn't have the goroutine?
func BufferFullDemo() bool {
	// YOUR CODE HERE
	return true
}

// ClosedChannelReceive demonstrates receiving from a closed channel.
//
// TODO: Implement this function to:
// 1. Create a buffered channel of string with capacity 2
// 2. Send "first" and "second" to the channel
// 3. Close the channel
// 4. Receive ALL values (including after close) and return them as a slice
//
// HINT: Use the "comma ok" idiom to detect when channel is exhausted
// QUESTION: How many receives can you do? What do you get after the buffered values?
func ClosedChannelReceive() []string {
	// YOUR CODE HERE
	return nil
}

// RangeOverChannel demonstrates using range to receive until close.
//
// TODO: Implement this function to:
// 1. Create an unbuffered channel of int
// 2. Launch a goroutine that sends values 1, 2, 3 then closes the channel
// 3. Use `for val := range ch` to collect all values
// 4. Return the sum of all values
//
// NOTE: range automatically stops when channel is closed
func RangeOverChannel() int {
	// YOUR CODE HERE
	return 0
}

// NilChannelBehavior demonstrates that nil channels block forever.
//
// TODO: Implement this function to:
// 1. Create a nil channel (var ch chan int, NOT make(chan int))
// 2. Use select with the nil channel and a timeout of 100ms
// 3. Return "timeout" if the select hit the timeout case
// 4. Return "received" if somehow a value was received (should never happen!)
//
// QUESTION: Why would you ever want a nil channel? (Hint: dynamic select cases)
func NilChannelBehavior() string {
	// YOUR CODE HERE
	return ""
}

// ChannelDirection demonstrates send-only and receive-only channel types.
// This is a compile-time safety feature.
//
// TODO: Complete these three functions:

// generator creates values and sends them on a send-only channel
func generator(out chan<- int, count int) {
	// YOUR CODE HERE
}

// squarer receives from one channel, squares, sends to another
func squarer(in <-chan int, out chan<- int) {
	// TODO: For each value from in, send its square to out
}

// ChannelDirectionDemo ties it together
// TODO: Create channels, wire up generator -> squarer, return sum of squares
func ChannelDirectionDemo(count int) int {
	// YOUR CODE HERE
	return 0
}

// =============================================================================
// CHALLENGE: Implement a timeout pattern without time.After
// =============================================================================

// SendWithTimeout attempts to send a value with a timeout.
// Returns true if send succeeded, false if timeout occurred.
//
// TODO: Implement WITHOUT using time.After (use time.NewTimer instead)
// Note: Go 1.23+ garbage collects unreferenced timers even if they haven't
// fired, but using time.NewTimer with Stop() is still good practice.
//
// HINT: time.NewTimer returns a *Timer with a channel C and method Stop()
func SendWithTimeout(ch chan<- int, value int, timeout time.Duration) bool {
	// YOUR CODE HERE
	return false
}
