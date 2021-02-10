package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestServer(t *testing.T) {
	fn := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "OK")
	}
	srv := NewServer()
	group := srv.RouteGroup("/kratos")
	{
		group.GET("/", fn)
		group.HEAD("/index", fn)
		group.OPTIONS("/home", fn)
		group.PUT("/products/{id}", fn)
		group.POST("/products/{id}", fn)
		group.PATCH("/products/{id}", fn)
		group.DELETE("/products/{id}", fn)
	}

	time.AfterFunc(time.Second, func() {
		srv.Stop(context.Background())
	})

	if err := srv.Start(context.Background()); !errors.Is(err, http.ErrServerClosed) {
		t.Fatal(err)
	}
}
