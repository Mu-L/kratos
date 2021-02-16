package http

import (
	"context"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport/http"
)

// Option is HTTP logging option.
type Option func(*options)

type options struct {
	logger log.Logger
}

// WithLogger with middleware logger.
func WithLogger(logger log.Logger) Option {
	return func(o *options) {
		o.logger = logger
	}
}

// Server is an HTTP logging middleware.
func Server(opts ...Option) middleware.Middleware {
	options := options{
		logger: log.DefaultLogger,
	}
	for _, o := range opts {
		o(&options)
	}
	log := log.NewHelper("logging/http", options.logger)
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			var (
				path   string
				method string
			)
			info, ok := http.FromContext(ctx)
			if ok {
				path = info.Request.RequestURI
				method = info.Request.Method
			}
			reply, err := handler(ctx, req)
			if err != nil {
				log.Errorw(
					"kind", "server",
					"http.path", path,
					"http.method", method,
					"http.code", errors.Code(err),
					"http.error", err.Error(),
				)
				return nil, err
			}
			log.Infow(
				"kind", "server",
				"http.path", path,
				"http.method", method,
				"http.code", 0,
			)
			return reply, nil
		}
	}
}
