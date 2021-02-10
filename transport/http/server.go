package http

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/log/stdlog"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"

	"github.com/gorilla/mux"
)

// SupportPackageIsVersion1 These constants should not be referenced from any other code.
const SupportPackageIsVersion1 = true

var _ transport.Server = (*Server)(nil)

// EncodeResponseFunc is encode response func.
type EncodeResponseFunc func(res http.ResponseWriter, req *http.Request, v interface{}) error

// EncodeErrorFunc is encode error func.
type EncodeErrorFunc func(res http.ResponseWriter, req *http.Request, err error)

// ServerOption is HTTP server option.
type ServerOption func(*serverOptions)

type serverOptions struct {
	network         string
	address         string
	timeout         time.Duration
	middleware      middleware.Middleware
	responseEncoder EncodeResponseFunc
	errorEncoder    EncodeErrorFunc
	logger          log.Logger
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

// Middleware with server middleware option.
func Middleware(m middleware.Middleware) ServerOption {
	return func(s *serverOptions) {
		s.middleware = m
	}
}

// ResponseEncoder with response handler option.
func ResponseEncoder(fn EncodeResponseFunc) ServerOption {
	return func(s *serverOptions) {
		s.responseEncoder = fn
	}
}

// ErrorEncoder with error handler option.
func ErrorEncoder(fn EncodeErrorFunc) ServerOption {
	return func(s *serverOptions) {
		s.errorEncoder = fn
	}
}

// Logger with server logger.
func Logger(logger log.Logger) ServerOption {
	return func(s *serverOptions) {
		s.logger = logger
	}
}

// Server is a HTTP server wrapper.
type Server struct {
	*http.Server
	opts   serverOptions
	router *mux.Router
	log    *log.Helper
}

// NewServer creates a HTTP server by options.
func NewServer(opts ...ServerOption) *Server {
	options := serverOptions{
		network:         "tcp",
		address:         ":8000",
		timeout:         time.Second,
		responseEncoder: DefaultResponseEncoder,
		errorEncoder:    DefaultErrorEncoder,
		logger:          stdlog.NewLogger(),
	}
	for _, o := range opts {
		o(&options)
	}
	srv := &Server{
		opts:   options,
		router: mux.NewRouter(),
		log:    log.NewHelper("http", options.logger),
	}
	srv.Server = &http.Server{Handler: srv}
	return srv
}

// RouteGroup .
func (s *Server) RouteGroup(path string) *RouteGroup {
	return &RouteGroup{root: path, router: s.router}
}

// Handle registers a new route with a matcher for the URL path.
func (s *Server) Handle(path string, h http.Handler) {
	s.router.Handle(path, h)
}

// HandleFunc registers a new route with a matcher for the URL path.
func (s *Server) HandleFunc(path string, h http.HandlerFunc) {
	s.router.HandleFunc(path, h)
}

// ServeHTTP should write reply headers and data to the ResponseWriter and then return.
func (s *Server) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithTimeout(req.Context(), s.opts.timeout)
	defer cancel()
	ctx = transport.NewContext(ctx, transport.Transport{Kind: "HTTP"})
	ctx = NewContext(ctx, ServerInfo{Request: req, Response: res})
	s.router.ServeHTTP(res, req.WithContext(ctx))
}

// Start start the HTTP server.
func (s *Server) Start(ctx context.Context) error {
	lis, err := net.Listen(s.opts.network, s.opts.address)
	if err != nil {
		return err
	}
	s.log.Infof("[HTTP] server listening on: %s", s.opts.address)
	return s.Serve(lis)
}

// Stop stop the HTTP server.
func (s *Server) Stop(ctx context.Context) error {
	s.log.Info("[HTTP] server stopping")
	return s.Shutdown(ctx)
}
