package gracewrap

import (
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"
)

func TestServeGRPC_StartsAndShutdowns(t *testing.T) {
	g := New(nil)
	g.config.HardStopTimeout = 0

	_, ln, err := g.ServeGRPC("127.0.0.1:0")
	if err != nil {
		t.Fatalf("serve grpc: %v", err)
	}
	if ln == nil {
		t.Fatalf("expected listener")
	}
	// let it start
	time.Sleep(30 * time.Millisecond)
	g.Shutdown()
}

func TestWrapGRPC_StartsAndShutdowns(t *testing.T) {
	g := New(nil)
	g.config.HardStopTimeout = 0

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()

	s := grpc.NewServer()
	if err := g.WrapGRPC(s, ln); err != nil {
		t.Fatalf("wrap grpc: %v", err)
	}
	// let it start
	time.Sleep(30 * time.Millisecond)
	g.Shutdown()
}
