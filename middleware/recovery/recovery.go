package recovery

import (
	"context"
	"runtime"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
)

// HandlerFunc is recovery handler func.
type HandlerFunc func(ctx context.Context, req, err interface{}) error

// Option is recovery option.
type Option func(*options)

type options struct {
	handler HandlerFunc
}

// Handler with recovery handler.
func Handler(h HandlerFunc) Option {
	return func(o *options) {
		o.handler = h
	}
}

// Recovery is a server middleware that recovers from any panics.
func Recovery(opts ...Option) middleware.Middleware {
	logger := log.NewHelper("middleware/recovery", log.DefaultLogger)
	options := options{
		handler: func(ctx context.Context, req, err interface{}) error {
			buf := make([]byte, 64<<10)
			n := runtime.Stack(buf, false)
			buf = buf[:n]
			logger.Errorf("%v: %+v\n%s\n", err, req, buf)
			return errors.Unknown("Unknown", "panic triggered: %v", err)
		},
	}
	for _, o := range opts {
		o(&options)
	}
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			defer func() {
				if rerr := recover(); rerr != nil {
					err = options.handler(ctx, req, rerr)
				}
			}()
			return handler(ctx, req)
		}
	}
}
