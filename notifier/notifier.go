package notifier

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	ErrQueueFull = errors.New("notifier queue full")
	ErrClosed    = errors.New("notifier closed")
)

type Result struct {
	Message string
	Err     error
	At      time.Time
}

type job struct {
	ctx     context.Context
	message string
	done    chan Result
}

// Client implements an asynchronous notification sender
type Client struct {
	queue  chan job
	wg     sync.WaitGroup
	mu     sync.Mutex
	closed bool
}

// New creates a notifier client with a bounded queue
// and a fixed number of workers
func New(workers, queueSize int) *Client {
	if workers <= 0 {
		workers = 8
	}
	if queueSize <= 0 {
		queueSize = 1024
	}

	c := &Client{
		queue: make(chan job, queueSize),
	}

	for i := 0; i < workers; i++ {
		c.wg.Add(1)
		go c.worker()
	}

	return c
}

func (c *Client) Notify(ctx context.Context, message string) <-chan Result {
	done := make(chan Result, 1)

	c.mu.Lock()
	closed := c.closed
	c.mu.Unlock()

	if closed {
		done <- Result{Message: message, Err: ErrClosed, At: time.Now()}
		return done
	}

	j := job{ctx: ctx, message: message, done: done}

	select {
	case c.queue <- j:
		return done
	default:
		done <- Result{Message: message, Err: ErrQueueFull, At: time.Now()}
		return done
	}
}

func (c *Client) Close(ctx context.Context) error {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return nil
	}
	c.closed = true
	close(c.queue)
	c.mu.Unlock()

	wait := make(chan struct{})
	go func() {
		c.wg.Wait()
		close(wait)
	}()

	select {
	case <-wait:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (c *Client) worker() {
	defer c.wg.Done()

	// Acknowledges messages but no HTTP delivery
	for j := range c.queue {
		j.done <- Result{Message: j.message, At: time.Now()}
	}
}
