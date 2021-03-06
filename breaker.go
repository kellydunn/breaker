package breaker

import (
	"errors"
	"sync"
	"sync/atomic"
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
	threshold uint64
	failures  uint64
	mu        sync.RWMutex
}

// NewBreaker creates and returns a pointer to a new Breaker
// with a failure threshold of the passed in value.
func NewBreaker(threshold uint64) *Breaker {
	return &Breaker{
		threshold: threshold,
		failures:  0,
	}
}

// IsOpen returns true if the current Breaker
// has a failure count above its failure threshold.
// Returns false otherwise.
func (b *Breaker) IsOpen() bool {
	return b.loadFailures() >= b.loadThreshold()
}

// IsClosed returns true if the current Breaker
// has a failure count below its failure threshold.
// Returns false otherwise.
func (b *Breaker) IsClosed() bool {
	return b.loadFailures() < b.loadThreshold()
}

func (b *Breaker) loadFailures() uint64 {
	return atomic.LoadUint64(&b.failures)
}

func (b *Breaker) loadThreshold() uint64 {
	return atomic.LoadUint64(&b.threshold)
}

// Trip increments the failure count of the current Breaker.
func (b *Breaker) Trip() {
	atomic.AddUint64(&b.failures, 1)
}

// Reset resets the current Breaker's failure count to zero.
func (b *Breaker) Reset() {
	atomic.StoreUint64(&b.failures, 0)
}

// Do calls the HandlerFunc associated with the current Breaker
// with the passed in arguments. Returns
func (b *Breaker) Do(f HandlerFunc, timeout time.Duration) error {
	timerChan := make(chan bool, 1)
	errChan := make(chan error, 1)

	defer close(timerChan)
	defer close(errChan)

	var once sync.Once
	var done sync.WaitGroup
	done.Add(1)

	// Setup a timer goroutine to
	// ensure the function runs within a timeout
	go func() {
		time.Sleep(timeout)

		once.Do(func() {
			timerChan <- true
			done.Done()
		})
	}()

	// Setup a goroutine that runs the desired function
	// and sends any error on the error channel
	go func() {
		err := f()

		once.Do(func() {
			if err != nil {
				errChan <- err
			}

			done.Done()
		})
	}()

	// Wait for either the timeout
	// or the function goroutine to complete
	done.Wait()

	var ret error

	select {
	case <-timerChan:
		b.Trip()
		ret = ErrTimeout

	case e := <-errChan:
		b.Trip()
		ret = e

	default:
		ret = nil
	}

	return ret
}
