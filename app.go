package kratos

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/log/stdlog"
	"github.com/go-kratos/kratos/v2/registry"
	"golang.org/x/sync/errgroup"
)

// App is an application components lifecycle manager
type App struct {
	opts   options
	log    *log.Helper
	cancel func()
}

// New create an application lifecycle manager.
func New(opts ...Option) *App {
	options := options{
		logger: stdlog.NewLogger(),
		ctx:    context.Background(),
		sigs:   []os.Signal{syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT},
	}
	for _, o := range opts {
		o(&options)
	}
	return &App{
		opts: options,
		log:  log.NewHelper("app", options.logger),
	}
}

// Service returns registry service.
func (a *App) Service() *registry.Service {
	return &registry.Service{
		ID:        a.opts.id,
		Name:      a.opts.name,
		Version:   a.opts.version,
		Metadata:  a.opts.metadata,
		Endpoints: a.opts.endpoints,
	}
}

// Run executes all OnStart hooks registered with the application's Lifecycle.
func (a *App) Run() error {
	var (
		ctx context.Context
	)
	ctx, a.cancel = context.WithCancel(a.opts.ctx)
	g, ctx := errgroup.WithContext(ctx)
	for _, srv := range a.opts.servers {
		srv := srv
		g.Go(func() error {
			<-ctx.Done() // wait for stop signal
			stopCtx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			return srv.Stop(stopCtx)
		})
		g.Go(func() error {
			startCtx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			return srv.Start(startCtx)
		})
	}
	if a.opts.registry != nil {
		g.Go(func() error {
			time.Sleep(time.Second) // wait for server started
			a.log.Infof("Registering %s service to the registry", a.opts.name)
			return a.opts.registry.Register(a.Service())
		})
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
func (a *App) Stop() {
	if a.opts.registry != nil {
		a.log.Infof("Unregistering in the registry service: %s", a.opts.name)
		if err := a.opts.registry.Deregister(a.Service()); err != nil {
			a.log.Errorf("Failed to deregister registry: %v", err)
		}
	}
	if a.cancel != nil {
		a.cancel()
	}
}
