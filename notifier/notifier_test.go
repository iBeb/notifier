package notifier

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNotify_ReturnsErrClosedAfterClose(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c := New(srv.URL, Options{
		Workers:        1,
		QueueSize:      1,
		RequestTimeout: 10 * time.Second,
	})

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
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	// 0 workers so nothing drains the queue, queue size 1.
	c := New(srv.URL, Options{
		Workers:        0,
		QueueSize:      1,
		RequestTimeout: 10 * time.Second,
	})
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = c.Close(ctx)
	}()

	_ = c.Notify(context.Background(), "msg1") // fills queue

	res := <-c.Notify(context.Background(), "msg2")
	if res.Err != ErrQueueFull {
		t.Fatalf("Err: expected %v, got %v", ErrQueueFull, res.Err)
	}
}

func TestNotify_IsNonBlocking(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	// no workers, but Notify must still return immediately
	c := New(srv.URL, Options{
		Workers:        0,
		QueueSize:      1,
		RequestTimeout: 10 * time.Second,
	})
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
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c := New(srv.URL, Options{
		Workers:        1,
		QueueSize:      10,
		RequestTimeout: 10 * time.Second,
	})
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

func TestWorker_SendsPOSTWithBodyAndContentType(t *testing.T) {
	var gotBody string
	var gotCT string
	var gotMethod string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotCT = r.Header.Get("Content-Type")
		b, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read body: %v", err)
		}
		gotBody = string(b)

		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c := New(srv.URL, Options{
		Workers:        1,
		QueueSize:      10,
		RequestTimeout: 10 * time.Second,
	})
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = c.Close(ctx)
	}()

	res := <-c.Notify(context.Background(), "hello")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if gotMethod != http.MethodPost {
		t.Fatalf("method: expected %s, got %s", http.MethodPost, gotMethod)
	}
	if gotBody != "hello" {
		t.Fatalf("body: expected %s, got %s", "hello", gotBody)
	}
	if gotCT != "text/plain; charset=utf-8" {
		t.Fatalf("content-type: got %q", gotCT)
	}
	if res.StatusCode != http.StatusNoContent {
		t.Fatalf("status: expected %d, got %d", http.StatusNoContent, res.StatusCode)
	}
}

func TestWorker_Non2xxReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := New(srv.URL, Options{
		Workers:        1,
		QueueSize:      10,
		RequestTimeout: 10 * time.Second,
	})
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = c.Close(ctx)
	}()

	res := <-c.Notify(context.Background(), "hello")
	if res.Err == nil {
		t.Fatalf("expected error, got nil")
	}
	if res.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: expected %d, got %d", http.StatusInternalServerError, res.StatusCode)
	}
}
