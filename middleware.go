package gracewrap

import (
	"context"
	"net/http"

	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

// httpMiddleware wraps an HTTP handler to track in-flight requests.
func (g *Graceful) httpMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		g.incInflight()
		defer g.decInflight()

		// Update metrics
		if g.metrics != nil {
			g.metrics.incHTTP()
		}

		next.ServeHTTP(w, r)
	})
}

// grpcUnaryInterceptor tracks in-flight unary RPCs.
func (g *Graceful) grpcUnaryInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	g.incInflight()
	defer g.decInflight()

	// Update metrics
	if g.metrics != nil {
		g.metrics.incGRPC()
	}

	return handler(ctx, req)
}

// grpcStreamInterceptor tracks in-flight streaming RPCs.
func (g *Graceful) grpcStreamInterceptor(
	srv interface{},
	ss grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	g.incInflight()
	defer g.decInflight()

	// Update metrics
	if g.metrics != nil {
		g.metrics.incGRPC()
	}

	return handler(srv, &trackedStream{ServerStream: ss, graceful: g})
}

// trackedStream wraps a gRPC ServerStream to track the connection.
type trackedStream struct {
	grpc.ServerStream
	graceful *Graceful
}

// RecvMsg implements the grpc.ServerStream interface.
func (ts *trackedStream) RecvMsg(m interface{}) error {
	return ts.ServerStream.RecvMsg(m)
}

// SendMsg implements the grpc.ServerStream interface.
func (ts *trackedStream) SendMsg(m interface{}) error {
	return ts.ServerStream.SendMsg(m)
}

// incInflight increments the in-flight request counter.
func (g *Graceful) incInflight() {
	g.inflight.mu.Lock()
	g.inflight.n++
	g.inflight.mu.Unlock()

	// Update metrics
	if g.metrics != nil {
		g.metrics.updateInflight(g.inflight.n)
	}
}

// decInflight decrements the in-flight request counter.
func (g *Graceful) decInflight() {
	g.inflight.mu.Lock()
	g.inflight.n--
	if g.inflight.n == 0 {
		g.inflight.cv.Broadcast()
	}
	g.inflight.mu.Unlock()

	// Update metrics
	if g.metrics != nil {
		g.metrics.updateInflight(g.inflight.n)
	}
}

// peerAddr extracts the peer address from a gRPC context.
// This is a helper function for logging/monitoring.
func peerAddr(ctx context.Context) string {
	if p, ok := peer.FromContext(ctx); ok && p.Addr != nil {
		return p.Addr.String()
	}
	return ""
}
