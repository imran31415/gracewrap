package gracewrap

import (
	"context"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

func TestMetricsGRPCIncrements(t *testing.T) {
	cfg := DefaultConfig()
	cfg.EnableMetrics = true
	cfg.PrometheusRegistry = prometheus.NewRegistry()
	g := New(&cfg)

	_, _ = g.grpcUnaryInterceptor(context.Background(), nil, nil, func(ctx context.Context, req interface{}) (interface{}, error) { return nil, nil })
	_ = g.grpcStreamInterceptor(nil, &fakeServerStream{}, nil, grpc.StreamHandler(func(srv interface{}, stream grpc.ServerStream) error { return nil }))
}

func TestPeerAddr(t *testing.T) {
	addr := "1.2.3.4:5678"
	p := &peer.Peer{Addr: &dummyAddr{network: "tcp", addr: addr}}
	ctx := peer.NewContext(context.Background(), p)
	if got := peerAddr(ctx); got == "" {
		t.Fatalf("expected peer address")
	}
}

type dummyAddr struct{ network, addr string }

func (d *dummyAddr) Network() string { return d.network }
func (d *dummyAddr) String() string  { return d.addr }
