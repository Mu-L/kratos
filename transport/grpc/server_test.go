package grpc

import (
	"testing"
	"time"
)

func TestServer(t *testing.T) {
	srv := NewServer()
	if endpoint, err := srv.Endpoint(); err != nil || endpoint == "" {
		t.Fatal(endpoint, err)
	}

	time.AfterFunc(time.Second, func() {
		srv.Stop()
	})

	if err := srv.Start(); err != nil {
		t.Fatal(err)
	}
}
