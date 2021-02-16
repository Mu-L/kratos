package http

import (
	"context"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/go-kratos/kratos/v2/encoding"
	"github.com/go-kratos/kratos/v2/errors"
)

// DecodeResponseFunc decode response func.
type DecodeResponseFunc func(res *http.Response, v interface{}) error

// ClientOption is HTTP client option.
type ClientOption func(*clientOptions)

// WithTimeout with client request timeout.
func WithTimeout(d time.Duration) ClientOption {
	return func(o *clientOptions) {
		o.timeout = d
	}
}

// WithUserAgent with client user agent.
func WithUserAgent(ua string) ClientOption {
	return func(o *clientOptions) {
		o.userAgent = ua
	}
}

// WithTransport with client transport.
func WithTransport(trans http.RoundTripper) ClientOption {
	return func(o *clientOptions) {
		o.transport = trans
	}
}

// Client is a HTTP transport client.
type clientOptions struct {
	timeout   time.Duration
	userAgent string
	transport http.RoundTripper
}

// NewClient returns an HTTP client.
func NewClient(ctx context.Context, opts ...ClientOption) (*http.Client, error) {
	trans, err := NewTransport(ctx, opts...)
	if err != nil {
		return nil, err
	}
	return &http.Client{Transport: trans}, nil
}

// NewTransport creates an http.RoundTripper.
func NewTransport(ctx context.Context, opts ...ClientOption) (http.RoundTripper, error) {
	options := &clientOptions{
		timeout:   500 * time.Millisecond,
		transport: http.DefaultTransport,
	}
	for _, o := range opts {
		o(options)
	}
	return &baseTransport{
		userAgent: options.userAgent,
		timeout:   options.timeout,
		base:      options.transport,
	}, nil
}

type baseTransport struct {
	userAgent string
	timeout   time.Duration
	base      http.RoundTripper
}

func (t *baseTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.userAgent != "" && req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", t.userAgent)
	}
	ctx, cancel := context.WithTimeout(req.Context(), t.timeout)
	defer cancel()
	res, err := t.base.RoundTrip(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	return res, nil
}

// CheckResponse returns an error (of type *Error) if the response
// status code is not 2xx.
func CheckResponse(res *http.Response) error {
	if res.StatusCode >= 200 && res.StatusCode <= 299 {
		return nil
	}
	se := &errors.StatusError{}
	if err := DecodeResponse(res, se); err != nil {
		return err
	}
	return se
}

// DecodeResponse decodes the body of res into target. If there is no body, target is unchanged.
func DecodeResponse(res *http.Response, v interface{}) error {
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	subtype := contentSubtype(res.Header.Get("content-type"))
	codec := encoding.GetCodec(subtype)
	if codec == nil {
		codec = encoding.GetCodec("json")
	}
	return codec.Unmarshal(data, v)
}
