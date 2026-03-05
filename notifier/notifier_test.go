package notifier

import (
	"context"
	"testing"
	"time"
)

func TestNotify_ReturnsErrClosedAfterClose(t *testing.T) {
	c := New(1, 1)

	ctxClose, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := c.Close(ctxClose); err != nil {
		t.Fatalf("Close: %v", err)
	}

	res := <-c.Notify(context.Background(), "hello")
	if res.Err != ErrClosed {
		t.Fatalf("Err: expected %v, got %v", ErrClosed, res.Err)
	}
}

func TestNotify_FailsFastWhenQueueIsFull(t *testing.T) {
	// 0 workers so nothing drains the queue, queue size 1.
	c := New(0, 1)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = c.Close(ctx)
	}()

	_ = c.Notify(context.Background(), "msg1") // fills queue

	res := <-c.Notify(context.Background(), "msg2")
	if res.Err != ErrQueueFull {
		t.Fatalf("Err: expected %v, got %v", ErrQueueFull, ErrClosed)
	}
}

func TestNotify_IsNonBlocking(t *testing.T) {
	c := New(0, 1) // no workers, but Notify must still return immediately
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = c.Close(ctx)
	}()

	start := time.Now()
	_ = c.Notify(context.Background(), "msg1")
	elapsed := time.Since(start)

	if elapsed > 50*time.Millisecond {
		t.Fatalf("Notify took too long: %s", elapsed)
	}
}

func TestNotify_AcknowledgesMessage(t *testing.T) {
	c := New(1, 10)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = c.Close(ctx)
	}()

	res := <-c.Notify(context.Background(), "hello")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if res.Message != "hello" {
		t.Fatalf("message: expected %q, got %q", "hello", res.Message)
	}
}
