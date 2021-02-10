package grpc

import (
	"context"
	"net"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/log/stdlog"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"

	"google.golang.org/grpc"
)

var _ transport.Server = (*Server)(nil)

// ServerOption is gRPC server option.
type ServerOption func(o *serverOptions)

type serverOptions struct {
	network     string
	address     string
	timeout     time.Duration
	interceptor grpc.UnaryServerInterceptor
	middleware  middleware.Middleware
	grpcOpts    []grpc.ServerOption
	logger      log.Logger
}

// Network with server network.
func Network(network string) ServerOption {
	return func(o *serverOptions) {
		o.network = network
	}
}

// Address with server address.
func Address(addr string) ServerOption {
	return func(o *serverOptions) {
		o.address = addr
	}
}

// Timeout with server timeout.
func Timeout(timeout time.Duration) ServerOption {
	return func(o *serverOptions) {
		o.timeout = timeout
	}
}

// UnaryInterceptor with server interceptor.
func UnaryInterceptor(in grpc.UnaryServerInterceptor) ServerOption {
	return func(o *serverOptions) {
		o.interceptor = in
	}
}

// Middleware with server middleware.
func Middleware(m middleware.Middleware) ServerOption {
	return func(o *serverOptions) {
		o.middleware = m
	}
}

// Logger with server logger.
func Logger(logger log.Logger) ServerOption {
	return func(o *serverOptions) {
		o.logger = logger
	}
}

// Options with grpc options.
func Options(opts ...grpc.ServerOption) ServerOption {
	return func(o *serverOptions) {
		o.grpcOpts = opts
	}
}

// Server is a gRPC server wrapper.
type Server struct {
	*grpc.Server
	opts serverOptions
	log  *log.Helper
}

// NewServer creates a gRPC server by options.
func NewServer(opts ...ServerOption) *Server {
	options := serverOptions{
		network: "tcp",
		address: ":9000",
		timeout: time.Second,
		logger:  stdlog.NewLogger(),
	}
	for _, o := range opts {
		o(&options)
	}
	var grpcOpts = []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			UnaryServerInterceptor(options.middleware),
			UnaryTimeoutInterceptor(options.timeout),
		),
	}
	if options.interceptor != nil {
		grpcOpts = append(grpcOpts, grpc.ChainUnaryInterceptor(
			options.interceptor,
			UnaryServerInterceptor(options.middleware),
			UnaryTimeoutInterceptor(options.timeout),
		))
	}
	if len(options.grpcOpts) > 0 {
		grpcOpts = append(grpcOpts, options.grpcOpts...)
	}
	return &Server{
		opts:   options,
		Server: grpc.NewServer(grpcOpts...),
		log:    log.NewHelper("grpc", options.logger),
	}
}

// Start start the gRPC server.
func (s *Server) Start() error {
	lis, err := net.Listen(s.opts.network, s.opts.address)
	if err != nil {
		return err
	}
	s.log.Infof("[gRPC] server listening on: %s", s.opts.address)
	return s.Serve(lis)
}

// Stop stop the gRPC server.
func (s *Server) Stop() error {
	s.GracefulStop()
	s.log.Info("[gRPC] server stopping")
	return nil
}

// UnaryTimeoutInterceptor returns a unary timeout interceptor.
func UnaryTimeoutInterceptor(timeout time.Duration) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		return handler(ctx, req)
	}
}

// UnaryServerInterceptor returns a unary server interceptor.
func UnaryServerInterceptor(m middleware.Middleware) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		ctx = transport.NewContext(ctx, transport.Transport{Kind: "GRPC"})
		ctx = NewContext(ctx, ServerInfo{Server: info.Server, FullMethod: info.FullMethod})
		h := func(ctx context.Context, req interface{}) (interface{}, error) {
			return handler(ctx, req)
		}
		if m != nil {
			h = m(h)
		}
		reply, err := h(ctx, req)
		if err != nil {
			return nil, err
		}
		return reply, nil
	}
}
