package http

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/go-kratos/kratos/v2/internal/host"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport"

	"github.com/gorilla/mux"
)

// SupportPackageIsVersion1 These constants should not be referenced from any other code.
const SupportPackageIsVersion1 = true

var logger = log.NewHelper(log.GetLogger("transport/http"))
var _ transport.Server = (*Server)(nil)

// DecodeRequestFunc deocder request func.
type DecodeRequestFunc func(req *http.Request, v interface{}) error

// EncodeResponseFunc is encode response func.
type EncodeResponseFunc func(res http.ResponseWriter, req *http.Request, v interface{}) error

// EncodeErrorFunc is encode error func.
type EncodeErrorFunc func(res http.ResponseWriter, req *http.Request, err error)

// ServerOption is HTTP server option.
type ServerOption func(*Server)

// Network with server network.
func Network(network string) ServerOption {
	return func(s *Server) {
		s.network = network
	}
}

// Address with server address.
func Address(addr string) ServerOption {
	return func(s *Server) {
		s.address = addr
	}
}

// Timeout with server timeout.
func Timeout(timeout time.Duration) ServerOption {
	return func(s *Server) {
		s.timeout = timeout
	}
}

// Middleware with server middleware option.
func Middleware(m middleware.Middleware) ServerOption {
	return func(s *Server) {
		s.middleware = m
	}
}

// RequestDecoder with request decoder option.
func RequestDecoder(fn DecodeRequestFunc) ServerOption {
	return func(s *Server) {
		s.requestDecoder = fn
	}
}

// ResponseEncoder with response handler option.
func ResponseEncoder(fn EncodeResponseFunc) ServerOption {
	return func(s *Server) {
		s.responseEncoder = fn
	}
}

// ErrorEncoder with error handler option.
func ErrorEncoder(fn EncodeErrorFunc) ServerOption {
	return func(s *Server) {
		s.errorEncoder = fn
	}
}

// Server is a HTTP server wrapper.
type Server struct {
	*http.Server
	lis             net.Listener
	network         string
	address         string
	timeout         time.Duration
	middleware      middleware.Middleware
	requestDecoder  DecodeRequestFunc
	responseEncoder EncodeResponseFunc
	errorEncoder    EncodeErrorFunc
	router          *mux.Router
}

// NewServer creates a HTTP server by options.
func NewServer(opts ...ServerOption) *Server {
	srv := &Server{
		network:         "tcp",
		address:         ":0",
		timeout:         time.Second,
		requestDecoder:  DefaultRequestDecoder,
		responseEncoder: DefaultResponseEncoder,
		errorEncoder:    DefaultErrorEncoder,
		middleware:      recovery.Recovery(),
	}
	for _, o := range opts {
		o(srv)
	}
	srv.router = mux.NewRouter()
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
	ctx, cancel := context.WithTimeout(req.Context(), s.timeout)
	defer cancel()
	ctx = transport.NewContext(ctx, transport.Transport{Kind: "HTTP"})
	ctx = NewContext(ctx, ServerInfo{Request: req, Response: res})
	s.router.ServeHTTP(res, req.WithContext(ctx))
}

// Start start the HTTP server.
func (s *Server) Start() error {
	lis, err := net.Listen(s.network, s.address)
	if err != nil {
		return err
	}
	s.lis = lis
	logger.Infof("[HTTP] server listening on: %s", lis.Addr().String())
	return s.Serve(lis)
}

// Stop stop the HTTP server.
func (s *Server) Stop() error {
	logger.Info("[HTTP] server stopping")
	return s.Shutdown(context.Background())
}

// Endpoint return a real address to registry endpoint.
// examples:
//   http://127.0.0.1:8000?isSecure=false
func (s *Server) Endpoint() (string, error) {
	addr, err := host.Extract(s.address, s.lis)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("http://%s", addr), err
}
