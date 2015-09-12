package breaker

import (
	"errors"
	"sync"
	"time"
)

// HandlerFunc is a function that gets called by a
// Breaker when the `Do` function is called.
// It accepts a list of arguments and returns an error
type HandlerFunc func() error

// ErrTimeout describes when a timeout threshold is exceeded.
var ErrTimeout = errors.New("Timeout execeed in circuit breaker.")

// Breaker is a struct that behaves similar
// to Martin Fowler's CircuitBreaker design pattern,
// but keeps track of failure counts and changes state accordingly.
// More here: http://martinfowler.com/bliki/CircuitBreaker.html
type Breaker struct {
	threshold int
	failures  int
	mu        sync.RWMutex
}

// NewBreaker creates and returns a pointer to a new Breaker
// with a failure threshold of the passed in value.
func NewBreaker(threshold int) *Breaker {
	return &Breaker{
		threshold: threshold,
		failures:  0,
	}
}

// IsOpen returns true if the current Breaker
// has a failure count above its failure threshold.
// Returns false otherwise.
func (b *Breaker) IsOpen() bool {
	var open bool

	b.mu.RLock()
	if b.failures >= b.threshold {
		open = true
	}
	b.mu.RUnlock()

	return open
}

// IsClosed returns true if the current Breaker
// has a failure count below its failure threshold.
// Returns false otherwise.
func (b *Breaker) IsClosed() bool {
	var closed bool

	b.mu.RLock()
	if b.failures < b.threshold {
		closed = true
	}
	b.mu.RUnlock()

	return closed
}

// Trip increments the failure count of the current Breaker.
func (b *Breaker) Trip() {
	b.mu.Lock()
	b.failures++
	b.mu.Unlock()
}

// Reset resets the current Breaker's failure count to zero.
func (b *Breaker) Reset() {
	b.mu.Lock()
	b.failures = 0
	b.mu.Unlock()
}

// Do calls the HandlerFunc associated with the current Breaker
// with the passed in arguments. Returns
func (b *Breaker) Do(f HandlerFunc, timeout time.Duration) error {
	timerChan := make(chan bool, 1)
	errChan := make(chan error, 1)
	successChan := make(chan bool, 1)

	defer close(timerChan)
	defer close(errChan)
	defer close(successChan)

	var done sync.WaitGroup

	go func() {
		time.Sleep(timeout)
		timerChan <- true

		b.mu.Lock()
		done.Add(1)
		done.Done()
		b.mu.Unlock()
	}()

	go func() {
		err := f()
		if err != nil {
			errChan <- err
		} else {
			successChan <- true
		}

		b.mu.Lock()
		done.Add(1)
		done.Done()
		b.mu.Unlock()
	}()

	done.Wait()

	var ret error

	select {
	case <-successChan:
		b.Reset()
		ret = nil

	case <-timerChan:
		b.Trip()
		ret = ErrTimeout

	case e := <-errChan:
		b.Trip()
		ret = e
	}

	return ret
}
