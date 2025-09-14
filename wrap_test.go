package gracewrap

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func TestWrapHTTP_StartsAndShutdowns(t *testing.T) {
	g := New(nil)

	mux := http.NewServeMux()
	mux.HandleFunc("/ok", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})
	srv := &http.Server{Addr: ":0", Handler: mux}

	if err := g.WrapHTTP(srv); err != nil {
		t.Fatalf("wrap http err: %v", err)
	}

	// Give it a moment to start (on random port, not asserting external conn)
	time.Sleep(50 * time.Millisecond)

	// Trigger shutdown quickly
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	go func() { _ = g.Wait(ctx) }()
	g.Shutdown()
}
