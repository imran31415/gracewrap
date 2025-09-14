package gracewrap

import (
	"context"
	"testing"

	"google.golang.org/grpc"
)

type testUnaryServer struct{}

func TestGRPCUnaryInterceptor(t *testing.T) {
	g := New(nil)
	g.config.HardStopTimeout = 0

	called := false
	h := func(ctx context.Context, req interface{}) (interface{}, error) {
		called = true
		return "ok", nil
	}

	resp, err := g.grpcUnaryInterceptor(context.Background(), "req", &grpc.UnaryServerInfo{FullMethod: "/svc/Method"}, h)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if resp.(string) != "ok" || !called {
		t.Fatalf("handler not called or wrong resp: %v", resp)
	}
}

type fakeServerStream struct{ grpc.ServerStream }

func (f *fakeServerStream) SendMsg(m interface{}) error { return nil }
func (f *fakeServerStream) RecvMsg(m interface{}) error { return nil }

func TestGRPCStreamInterceptor(t *testing.T) {
	g := New(nil)
	g.config.HardStopTimeout = 0

	called := false
	h := func(srv interface{}, stream grpc.ServerStream) error {
		called = true
		return nil
	}

	if err := g.grpcStreamInterceptor(&testUnaryServer{}, &fakeServerStream{}, &grpc.StreamServerInfo{FullMethod: "/svc/Stream"}, h); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !called {
		t.Fatalf("stream handler not called")
	}
}
