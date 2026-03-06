package notifier

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"
)

var (
	ErrQueueFull = errors.New("notifier queue full")
	ErrClosed    = errors.New("notifier closed")
)

type Result struct {
	Message    string
	Err        error
	StatusCode int
	At         time.Time
}

type job struct {
	ctx     context.Context
	message string
	done    chan Result
}

// Client implements an asynchronous notification sender
type Client struct {
	url    string
	client *http.Client
	queue  chan job
	wg     sync.WaitGroup
	mu     sync.Mutex
	closed bool
}

// New creates a notifier client with a bounded queue
// and a fixed number of workers
func New(url string, opts Options) *Client {
    // Apply default values if needed
    opts = opts.withDefaults()

	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          256,
		MaxIdleConnsPerHost:   64,
		MaxConnsPerHost:       opts.Workers, // cap concurrency at transport level too
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   5 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ForceAttemptHTTP2:     true,
	}

	httpClient := &http.Client{
		Transport: transport,
		Timeout:   opts.RequestTimeout,
	}

	c := &Client{
		url:    url,
		client: httpClient,
		queue:  make(chan job, opts.QueueSize),
	}

	for i := 0; i < opts.Workers; i++ {
		c.wg.Add(1)
		go c.worker()
	}

	return c
}

func (c *Client) Notify(ctx context.Context, message string) <-chan Result {
	done := make(chan Result, 1)

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
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

	for j := range c.queue {
		res := Result{Message: j.message, At: time.Now()}

		req, err := http.NewRequestWithContext(j.ctx, http.MethodPost, c.url, bytes.NewBufferString(j.message))
		if err != nil {
			res.Err = err
			j.done <- res
			continue
		}
		req.Header.Set("Content-Type", "text/plain; charset=utf-8")

		resp, err := c.client.Do(req)
		if err != nil {
			res.Err = err
			j.done <- res
			continue
		}

		_ = resp.Body.Close()

		res.StatusCode = resp.StatusCode
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			res.Err = fmt.Errorf("non-2xx response: %d", resp.StatusCode)
		}

		j.done <- res
	}
}
