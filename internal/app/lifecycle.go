package app

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// Component.Start must honor its context and clean its own partially initialized
// resources on error. Stop must honor its context. A lifecycle is single-use.
type Component interface {
	Start(context.Context) error
	Stop(context.Context) error
}
type Lifecycle struct {
	mu             sync.Mutex
	started        []Component
	cancel         context.CancelFunc
	used, stopping bool
	stopErr        error
	stopDone       chan struct{}
}

func (l *Lifecycle) Start(ctx context.Context, components ...Component) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.used {
		return errors.New("lifecycle already used")
	}
	l.used = true
	l.stopDone = make(chan struct{})
	child, cancel := context.WithCancel(ctx)
	l.cancel = cancel
	if err := child.Err(); err != nil {
		return err
	}
	for _, c := range components {
		if err := child.Err(); err != nil {
			return err
		}
		if err := c.Start(child); err != nil {
			return fmt.Errorf("component initialization failed: %w", err)
		}
		l.started = append(l.started, c)
	}
	return nil
}
func (l *Lifecycle) Stop(ctx context.Context) error {
	l.mu.Lock()
	if !l.used {
		l.mu.Unlock()
		return nil
	}
	done := l.stopDone
	if !l.stopping {
		l.stopping = true
		components := append([]Component(nil), l.started...)
		cancel := l.cancel
		go func() {
			cancel()
			var result error
			for i := len(components) - 1; i >= 0; i-- {
				finished := make(chan error, 1)
				go func(c Component) { finished <- c.Stop(ctx) }(components[i])
				select {
				case e := <-finished:
					result = errors.Join(result, e)
				case <-ctx.Done():
					result = errors.Join(result, ctx.Err())
				}
			}
			l.mu.Lock()
			l.stopErr = result
			l.mu.Unlock()
			close(done)
		}()
	}
	l.mu.Unlock()
	select {
	case <-done:
		l.mu.Lock()
		defer l.mu.Unlock()
		return l.stopErr
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Run cleans partial startup, propagates fatal worker errors and bounds shutdown.
func Run(ctx context.Context, timeout time.Duration, fatal <-chan error, components ...Component) error {
	l := new(Lifecycle)
	err := l.Start(ctx, components...)
	if err == nil {
		select {
		case <-ctx.Done():
		case err = <-fatal:
		}
	}
	stopCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return errors.Join(err, l.Stop(stopCtx))
}
