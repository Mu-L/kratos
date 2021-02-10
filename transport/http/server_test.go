package http

import (
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
		srv.Stop()
	})

	if err := srv.Start(); !errors.Is(err, http.ErrServerClosed) {
		t.Fatal(err)
	}
}
