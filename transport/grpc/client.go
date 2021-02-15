package grpc

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/go-kratos/kratos/v2/transport/grpc/resolver/discovery"

	"google.golang.org/grpc"
)

// ClientOption is gRPC client option.
type ClientOption func(o *Client)

// WithContext with client context.
func WithContext(ctx context.Context) ClientOption {
	return func(c *Client) {
		c.ctx = ctx
	}
}

// WithTimeout with client timeout.
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.timeout = timeout
	}
}

// WithInsecure with client insecure.
func WithInsecure() ClientOption {
	return func(c *Client) {
		c.insecure = true
	}
}

// WithMiddleware with client middleware.
func WithMiddleware(m middleware.Middleware) ClientOption {
	return func(c *Client) {
		c.middleware = m
	}
}

// WithRegistry with client registry.
func WithRegistry(r registry.Registry) ClientOption {
	return func(c *Client) {
		c.registry = r
	}
}

// WithOptions with gRPC options.
func WithOptions(opts ...grpc.DialOption) ClientOption {
	return func(c *Client) {
		c.grpcOpts = opts
	}
}

// Client is gRPC Client
type Client struct {
	*grpc.ClientConn
	ctx        context.Context
	insecure   bool
	timeout    time.Duration
	middleware middleware.Middleware
	registry   registry.Registry
	grpcOpts   []grpc.DialOption
}

// NewClient new a grpc transport client.
func NewClient(target string, opts ...ClientOption) (client *Client, err error) {
	client = &Client{
		ctx:        context.Background(),
		timeout:    500 * time.Millisecond,
		insecure:   false,
		middleware: recovery.Recovery(),
	}
	for _, o := range opts {
		o(client)
	}
	var grpcOpts = []grpc.DialOption{
		grpc.WithTimeout(client.timeout),
		grpc.WithUnaryInterceptor(UnaryClientInterceptor(client.middleware)),
	}
	if client.registry != nil {
		grpc.WithResolvers(discovery.NewBuilder(client.registry))
	}
	if client.insecure {
		grpcOpts = append(grpcOpts, grpc.WithInsecure())
	}
	if len(client.grpcOpts) > 0 {
		grpcOpts = append(grpcOpts, client.grpcOpts...)
	}
	if client.ClientConn, err = grpc.DialContext(client.ctx, target, grpcOpts...); err != nil {
		return
	}
	return
}

// UnaryClientInterceptor retruns a unary client interceptor.
func UnaryClientInterceptor(m middleware.Middleware) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		h := func(ctx context.Context, req interface{}) (interface{}, error) {
			return reply, invoker(ctx, method, req, reply, cc, opts...)
		}
		if m != nil {
			h = m(h)
		}
		_, err := h(ctx, req)
		return err
	}
}
