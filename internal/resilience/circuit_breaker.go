package resilience

import (
	"context"
	"time"

	"github.com/sony/gobreaker/v2"
)

// CircuitBreaker обёртка над gobreaker с указанием типа interface{}
type CircuitBreaker struct {
	cb *gobreaker.CircuitBreaker[interface{}]
}

func NewCircuitBreaker(name string, maxRequests uint32, interval, timeout time.Duration) *CircuitBreaker {
	cb := gobreaker.NewCircuitBreaker[interface{}](gobreaker.Settings{
		Name:        name,
		MaxRequests: maxRequests,
		Interval:    interval,
		Timeout:     timeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures > 5
		},
	})
	return &CircuitBreaker{cb: cb}
}

func (cb *CircuitBreaker) Execute(fn func() error) error {
	_, err := cb.cb.Execute(func() (interface{}, error) {
		return nil, fn()
	})
	return err
}

func (cb *CircuitBreaker) ExecuteContext(ctx context.Context, fn func() error) error {
	errCh := make(chan error, 1)
	go func() {
		_, err := cb.cb.Execute(func() (interface{}, error) {
			return nil, fn()
		})
		errCh <- err
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errCh:
		return err
	}
}