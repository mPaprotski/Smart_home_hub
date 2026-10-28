package app

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"
)

type component struct {
	id    int
	trace *[]int
	fail  bool
}

func (c component) Start(context.Context) error {
	*c.trace = append(*c.trace, c.id)
	if c.fail {
		return errors.New("failure")
	}
	return nil
}
func (c component) Stop(context.Context) error { *c.trace = append(*c.trace, -c.id); return nil }
func TestStartupFailureCleansInReverse(t *testing.T) {
	trace := []int{}
	err := Run(context.Background(), time.Second, nil, component{1, &trace, false}, component{2, &trace, false}, component{3, &trace, true})
	if err == nil || !reflect.DeepEqual(trace, []int{1, 2, 3, -2, -1}) {
		t.Fatal(trace, err)
	}
}
func TestCancellationAndIdempotentStop(t *testing.T) {
	trace := []int{}
	l := new(Lifecycle)
	if err := l.Start(context.Background(), component{1, &trace, false}); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		if err := l.Stop(context.Background()); err != nil {
			t.Fatal(err)
		}
	}
	if !reflect.DeepEqual(trace, []int{1, -1}) {
		t.Fatal(trace)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := Run(ctx, time.Second, nil); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}

type waiting struct{}

func (waiting) Start(context.Context) error    { return nil }
func (waiting) Stop(ctx context.Context) error { <-ctx.Done(); return ctx.Err() }
func TestShutdownDeadline(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	fatal := make(chan error, 1)
	fatal <- errors.New("worker failure")
	start := time.Now()
	err := Run(ctx, 20*time.Millisecond, fatal, waiting{})
	cancel()
	if !errors.Is(err, context.DeadlineExceeded) || time.Since(start) > time.Second {
		t.Fatal(err)
	}
}
