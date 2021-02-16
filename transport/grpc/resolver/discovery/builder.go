package discovery

import (
	"github.com/go-kratos/kratos/v2/registry"
	"google.golang.org/grpc/resolver"
)

const name = "discovery"

type builder struct {
	registry registry.Registry
}

// NewBuilder creates a builder which is used to factory registry resolvers.
func NewBuilder(r registry.Registry) resolver.Builder {
	return &builder{
		registry: r,
	}
}

func (d *builder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	w, err := d.registry.Watch(target.Endpoint)
	if err != nil {
		return nil, err
	}
	r := &discoveryResolver{
		w:  w,
		cc: cc,
	}
	go r.watch()
	return r, nil
}

func (d *builder) Scheme() string {
	return name
}
