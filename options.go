package kratos

import (
	"os"
	"time"

	"github.com/go-kratos/kratos/v2/registry"
)

// Option is an application option.
type Option func(o *options)

// options is an application options.
type options struct {
	id        string
	name      string
	version   string
	metadata  map[string]string
	endpoints []string

	registry registry.Registry

	startTimeout time.Duration
	stopTimeout  time.Duration

	sigs  []os.Signal
	sigFn func(*App, os.Signal)
}

// ID with service id.
func ID(id string) Option {
	return func(o *options) { o.id = id }
}

// Name with service name.
func Name(name string) Option {
	return func(o *options) { o.name = name }
}

// Version with service version.
func Version(version string) Option {
	return func(o *options) { o.version = version }
}

// Metadata with service metadata.
func Metadata(md map[string]string) Option {
	return func(o *options) { o.metadata = md }
}

// Endpoints with service endpoint.
func Endpoints(endpoints []string) Option {
	return func(o *options) { o.endpoints = endpoints }
}

// Registry with service registry.
func Registry(r registry.Registry) Option {
	return func(o *options) { o.registry = r }
}

// StartTimeout with start timeout.
func StartTimeout(d time.Duration) Option {
	return func(o *options) { o.startTimeout = d }
}

// StopTimeout with stop timeout.
func StopTimeout(d time.Duration) Option {
	return func(o *options) { o.stopTimeout = d }
}

// Signal with os signals.
func Signal(fn func(*App, os.Signal), sigs ...os.Signal) Option {
	return func(o *options) {
		o.sigFn = fn
		o.sigs = sigs
	}
}
