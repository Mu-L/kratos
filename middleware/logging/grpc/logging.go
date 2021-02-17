package grpc

import (
	"context"
	"path"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport/grpc"
)

// Option is gRPC logging option.
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

// Server is a gRPC logging middleware.
func Server(opts ...Option) middleware.Middleware {
	options := options{
		logger: log.DefaultLogger,
	}
	for _, o := range opts {
		o(&options)
	}
	log := log.NewHelper("middleware/grpc", options.logger)
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			var (
				fullMethod string
				service    string
				method     string
			)
			info, ok := grpc.FromContext(ctx)
			if ok {
				fullMethod = info.FullMethod
				service = path.Dir(info.FullMethod)[1:]
				method = path.Base(info.FullMethod)
			}
			reply, err := handler(ctx, req)
			if err != nil {
				log.Errorw(
					"kind", "server",
					"grpc.path", fullMethod,
					"grpc.service", service,
					"grpc.method", method,
					"grpc.code", errors.Code(err),
					"grpc.error", err.Error(),
				)
				return nil, err
			}
			log.Infow(
				"kind", "server",
				"grpc.path", fullMethod,
				"grpc.service", service,
				"grpc.method", method,
				"grpc.code", 0,
			)
			return reply, nil
		}
	}
}
