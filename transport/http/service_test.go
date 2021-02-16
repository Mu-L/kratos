package http

import (
	"context"
	"net/http"
	"testing"

	"github.com/go-kratos/kratos/v2/middleware"
)

type testRequest struct{}
type testReply struct{}
type testService struct {
}

func (s *testService) SayHello(context.Context, interface{}) (interface{}, error) {
	return nil, nil
}

func TestService(t *testing.T) {
	h := func(srv interface{}, ctx context.Context, req *http.Request, dec func(interface{}) error, m middleware.Middleware) (interface{}, error) {
		var in testRequest
		if err := dec(&in); err != nil {
			return nil, err
		}
		h := func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.(*testService).SayHello(ctx, req)
		}
		out, err := m(h)(ctx, &in)
		if err != nil {
			return nil, err
		}
		return out, nil
	}
	sd := &ServiceDesc{
		ServiceName: "helloworld.Greeter",
		Methods: []MethodDesc{
			{
				Path:    "/helloworld",
				Method:  "GET",
				Handler: h,
			},
		},
	}

	svc := &testService{}
	srv := NewServer()
	srv.RegisterService(sd, svc)
}
