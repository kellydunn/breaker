package breaker

import (
	"errors"
	"sync"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	b := NewBreaker(1)

	if b.threshold != 1 {
		t.Errorf("Unexpected threshold for new breaker: %d", b.threshold)
	}

	if b.failures != 0 {
		t.Errorf("Unexpected count for new breaker: %d", b.failures)
	}

	if b.IsOpen() {
		t.Errorf("A new breaker should not be open")
	}
}

func TestIsOpen(t *testing.T) {
	b := NewBreaker(1)

	if b.IsOpen() {
		t.Errorf("A new breaker should not be open")
	}

	b.Trip()

	if !b.IsOpen() {
		t.Errorf("A Tripped breaker with a threshold of 1 should be open")
	}
}

func TestIsClosed(t *testing.T) {
	b := NewBreaker(1)

	if !b.IsClosed() {
		t.Errorf("A new breaker should be closed")
	}

	b.Trip()

	if b.IsClosed() {
		t.Errorf("A Tripped breaker with a threshold of 1 should not be closed")
	}
}

func TestTrip(t *testing.T) {
	b := NewBreaker(1)

	b.Trip()
	if b.failures != 1 {
		t.Errorf("A Tripped breaker that was just made should have a count of 1")
	}
}

func TestReset(t *testing.T) {
	b := NewBreaker(1)

	b.Trip()
	b.Reset()
	if b.failures != 0 {
		t.Errorf("A Tripped and then Reset breaker that was just made should have a count of 0")
	}

	if b.IsOpen() {
		t.Errorf("A freshly reset breaker should not be open.")
	}
}

func TestDo(t *testing.T) {
	b := NewBreaker(1)
	b.Do(func() error {
		return errors.New("Test Error")
	}, 0)

	if !b.IsOpen() {
		t.Errorf("Expected a function that throws an error to close the breaker")
	}

	b2 := NewBreaker(1)
	b2.Do(func() error {
		time.Sleep(time.Second)
		return nil
	}, 0)

	if !b2.IsOpen() {
		t.Errorf("Expected a function that times out to trip the breaker")
	}

	b3 := NewBreaker(1)
	err := b3.Do(func() error {
		return nil
	}, time.Second)

	if err != nil {
		t.Errorf("Did not expect an error for a successful call to Do.")
	}
}

func TestTripAsync(t *testing.T) {
	b := NewBreaker(10)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		i := 0
		for i < 5 {
			time.Sleep(time.Microsecond)
			b.Trip()
			i++
		}
		wg.Done()
	}()

	go func() {
		i := 0
		for i < 5 {
			time.Sleep(time.Microsecond)
			b.Trip()
			i++
		}

		wg.Done()
	}()

	wg.Wait()

	if b.IsClosed() {
		t.Errorf("Breaker should not be closed.")
	}
}
