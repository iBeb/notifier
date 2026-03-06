package notifier

import (
	"testing"
	"time"
)

func TestOptions_withDefaults(t *testing.T) {
	got := (Options{}).withDefaults()

	if got.Workers != 0 {
		t.Fatalf("Workers: got %d, want %d", got.Workers, 0)
	}
	if got.QueueSize != 1024 {
		t.Fatalf("QueueSize: got %d, want %d", got.QueueSize, 1024)
	}
	if got.RequestTimeout != 10*time.Second {
		t.Fatalf("RequestTimeout: got %s, want %s", got.RequestTimeout, 10*time.Second)
	}
}

func TestOptions_withDefaults_DoesNotOverride(t *testing.T) {
	want := Options{
		Workers:        3,
		QueueSize:      7,
		RequestTimeout: 250 * time.Millisecond,
	}
	got := want.withDefaults()

	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}
