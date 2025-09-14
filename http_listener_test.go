package gracewrap

import (
	"net"
	"net/http"
	"testing"
	"time"
)

func TestWrapHTTPWithListener(t *testing.T) {
	g := New(nil)
	g.config.HardStopTimeout = 0

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	srv := &http.Server{Handler: mux}

	if err := g.WrapHTTPWithListener(srv, ln); err != nil {
		t.Fatalf("wrap with listener: %v", err)
	}
	// let it start
	time.Sleep(30 * time.Millisecond)
	// shutdown promptly
	g.Shutdown()
}
