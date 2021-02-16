package kratos

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-kratos/kratos/v2/registry"
	"golang.org/x/sync/errgroup"
)

// App is an application components lifecycle manager
type App struct {
	opts   options
	ctx    context.Context
	cancel func()
}

// New create an application lifecycle manager.
func New(opts ...Option) *App {
	options := options{
		ctx:  context.Background(),
		sigs: []os.Signal{syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT},
	}
	for _, o := range opts {
		o(&options)
	}
	ctx, cancel := context.WithCancel(options.ctx)
	return &App{
		ctx:    ctx,
		cancel: cancel,
		opts:   options,
	}
}

func (a *App) serviceInstance() *registry.ServiceInstance {
	if len(a.opts.endpoints) == 0 {
		for _, srv := range a.opts.servers {
			if e, err := srv.Endpoint(); err == nil {
				a.opts.endpoints = append(a.opts.endpoints, e)
			}
		}
	}
	return &registry.ServiceInstance{
		ID:        a.opts.id,
		Name:      a.opts.name,
		Version:   a.opts.version,
		Metadata:  a.opts.metadata,
		Endpoints: a.opts.endpoints,
	}
}

// Run executes all OnStart hooks registered with the application's Lifecycle.
func (a *App) Run() error {
	g, ctx := errgroup.WithContext(a.ctx)
	for _, srv := range a.opts.servers {
		srv := srv
		g.Go(func() error {
			<-ctx.Done() // wait for stop signal
			return srv.Stop()
		})
		g.Go(func() error {
			return srv.Start()
		})
	}
	if a.opts.registry != nil {
		if err := a.opts.registry.Register(a.serviceInstance()); err != nil {
			return err
		}
	}
	c := make(chan os.Signal, 1)
	signal.Notify(c, a.opts.sigs...)
	g.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-c:
				a.Stop()
			}
		}
	})
	if err := g.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		return err
	}
	return nil
}

// Stop gracefully stops the application.
func (a *App) Stop() error {
	if a.opts.registry != nil {
		if err := a.opts.registry.Deregister(a.serviceInstance()); err != nil {
			return err
		}
	}
	if a.cancel != nil {
		a.cancel()
	}
	return nil
}
