package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/iBeb/notifier/notifier"
)

var resultWG sync.WaitGroup
var mu sync.Mutex

func main() {
	url := flag.String("url", "", "notification URL")
	interval := flag.Duration("interval", 5*time.Second, "notification interval")
	workers := flag.Int("workers", 8, "worker concurrency")
	queue := flag.Int("queue", 1024, "queue size")
	timeout := flag.Duration("timeout", 10*time.Second, "request timeout")
	flag.Parse()

	if *url == "" {
		fmt.Fprintln(os.Stderr, "--url is required")
		os.Exit(2)
	}

	n := notifier.New(*url, *workers, *queue, *timeout)

	lines := make(chan string, 1024)
	go func() {
		sc := bufio.NewScanner(os.Stdin)
		for sc.Scan() {
			lines <- sc.Text()
		}
		close(lines)
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(*interval)
	defer ticker.Stop()

	var sent, okCount, failCount int

loop:
	for {
		select {
		case <-sig:
			fmt.Fprintln(os.Stderr, "SIGINT received, shutting down...")
			break loop
		case <-ticker.C:
			select {
			case msg, more := <-lines:
				if !more {
					break loop
				}

				sent++
				ctx, cancel := context.WithTimeout(context.Background(), *timeout)
				resCh := n.Notify(ctx, msg)

				resultWG.Add(1)
				go func() {
					defer resultWG.Done()
					defer cancel()

					res := <-resCh

					mu.Lock()
					defer mu.Unlock()

					if res.Err != nil {
						failCount++
						fmt.Fprintf(os.Stderr, "FAIL: %q: %v\n", res.Message, res.Err)
					} else {
						okCount++
					}
				}()
			default:
				// No input available yet; just wait for next tick or a signal.
			}
		}
	}

	// Graceful shutdown
	closeCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = n.Close(closeCtx)
	resultWG.Wait()

	fmt.Fprintf(os.Stderr, "sent=%d ok=%d fail=%d\n", sent, okCount, failCount)
}