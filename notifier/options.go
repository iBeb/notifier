package notifier

import "time"

type Options struct {
	Workers        int
	QueueSize      int
	RequestTimeout time.Duration
}

func (o Options) withDefaults() Options {
	if o.Workers < 0 {
		o.Workers = 8
	}
	if o.QueueSize <= 0 {
		o.QueueSize = 1024
	}
	if o.RequestTimeout <= 0 {
		o.RequestTimeout = 10 * time.Second
	}
	return o
}
