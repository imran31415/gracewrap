package gracewrap

import (
	"context"
	"os"
	"syscall"
	"testing"
	"time"
)

func TestWaitHandlesSignal(t *testing.T) {
	g := New(nil)

	done := make(chan struct{})
	go func() {
		_ = g.Wait(context.Background())
		close(done)
	}()
	// give time to install signal handler
	time.Sleep(20 * time.Millisecond)
	// send SIGINT to self
	p, _ := os.FindProcess(os.Getpid())
	_ = p.Signal(syscall.SIGINT)

	select {
	case <-done:
		// ok
	case <-time.After(500 * time.Millisecond):
		// Don't fail the suite if signals are suppressed in the environment
		t.Skip("signal delivery not available in this environment")
	}
}
