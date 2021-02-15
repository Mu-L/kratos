package discovery

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/log/stdlog"
	"github.com/go-kratos/kratos/v2/registry"
	"google.golang.org/grpc/resolver"
)

const name = "discovery"

// Option is discovery option.
type Option func(*builder)

// WithLogger with discovery logger.
func WithLogger(l log.Logger) Option {
	return func(b *builder) {
		b.log = log.NewHelper(name, l)
	}
}

type builder struct {
	registry registry.Registry
	log      *log.Helper
}

// NewBuilder creates a builder which is used to factory registry resolvers.
func NewBuilder(r registry.Registry, opts ...Option) resolver.Builder {
	b := &builder{
		registry: r,
		log:      log.NewHelper(name, stdlog.NewLogger()),
	}
	for _, o := range opts {
		o(b)
	}
	return b
}

func (d *builder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	w, err := d.registry.Watch(target.Endpoint)
	if err != nil {
		return nil, err
	}
	r := &discoveryResolver{
		w:   w,
		cc:  cc,
		log: d.log,
	}
	go r.watch()
	return r, nil
}

func (d *builder) Scheme() string {
	return name
}
