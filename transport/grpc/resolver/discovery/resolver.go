package discovery

import (
	"net/url"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/registry"
	"google.golang.org/grpc/resolver"
)

type discoveryResolver struct {
	w   registry.Watcher
	cc  resolver.ClientConn
	log *log.Helper
}

func (r *discoveryResolver) watch() {
	for {
		updated, err := r.w.Next()
		if err != nil {
			r.log.Errorf("Failed to watch discovery endpoint: %v", err)
			time.Sleep(time.Second)
			continue
		}
		r.update(updated)
	}
}

func (r *discoveryResolver) update(updated []*registry.ServiceInstance) {
	var addrs []resolver.Address
	for _, up := range updated {
		endpoint, err := parseEndpoint(up.Endpoints)
		if err != nil {
			r.log.Errorf("Failed to parse discovery endpoint: %v", err)
			continue
		}
		addr := resolver.Address{
			Addr:     endpoint,
			Metadata: up.Metadata,
		}
		addrs = append(addrs, addr)
	}
	r.cc.UpdateState(resolver.State{Addresses: addrs})
}

func (r *discoveryResolver) Close() {
	r.w.Close()
}

func (r *discoveryResolver) ResolveNow(options resolver.ResolveNowOptions) {}

func parseEndpoint(endpoints []string) (string, error) {
	for _, e := range endpoints {
		u, err := url.Parse(e)
		if err != nil {
			return "", err
		}
		if u.Scheme == "grpc" {
			return u.Path, nil
		}
	}
	return "", nil
}
