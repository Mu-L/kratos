package http

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/go-kratos/kratos/v2/errors"
)

// DecodeResponseFunc decode response func.
type DecodeResponseFunc func(res *http.Response, v interface{}) error

// ClientOption is HTTP client option.
type ClientOption func(*Client)

// WithTimeout with client request timeout.
func WithTimeout(d time.Duration) ClientOption {
	return func(c *Client) {
		c.timeout = d
	}
}

// WithKeepAlive with client keepavlie.
func WithKeepAlive(d time.Duration) ClientOption {
	return func(c *Client) {
		c.keepAlive = d
	}
}

// WithMaxIdleConns with client max idle conns.
func WithMaxIdleConns(n int) ClientOption {
	return func(c *Client) {
		c.maxIdleConns = n
	}
}

// WithUserAgent with client user agent.
func WithUserAgent(ua string) ClientOption {
	return func(c *Client) {
		c.userAgent = ua
	}
}

// WithResponseDecoder with response decoder.
func WithResponseDecoder(d DecodeResponseFunc) ClientOption {
	return func(c *Client) {
		c.decoder = d
	}
}

// Client is a HTTP transport client.
type Client struct {
	base         *http.Client
	round        http.RoundTripper
	timeout      time.Duration
	keepAlive    time.Duration
	maxIdleConns int
	userAgent    string
	decoder      DecodeResponseFunc
}

// NewClient new a HTTP transport client.
func NewClient(opts ...ClientOption) (*Client, error) {
	client := &Client{
		timeout:      500 * time.Millisecond,
		keepAlive:    30 * time.Second,
		maxIdleConns: 100,
		decoder:      DefaultResponseDecoder,
	}
	for _, o := range opts {
		o(client)
	}
	client.round = &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   client.timeout,
			KeepAlive: client.keepAlive,
			DualStack: true,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          client.maxIdleConns,
		MaxIdleConnsPerHost:   client.maxIdleConns,
		IdleConnTimeout:       client.keepAlive,
		TLSHandshakeTimeout:   client.timeout,
		ExpectContinueTimeout: client.timeout,
	}
	client.base = &http.Client{Transport: client}
	return client, nil
}

// Do sends an HTTP request and returns an HTTP response.
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	return c.base.Do(req)
}

// RoundTrip is transport round trip.
func (c *Client) RoundTrip(req *http.Request) (*http.Response, error) {
	if c.userAgent != "" && req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", c.userAgent)
	}
	ctx, cancel := context.WithTimeout(req.Context(), c.timeout)
	defer cancel()
	res, err := c.round.RoundTrip(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	return res, nil
}

// CheckResponse returns an error (of type *Error) if the response
// status code is not 2xx.
func (c *Client) CheckResponse(res *http.Response) error {
	if res.StatusCode >= 200 && res.StatusCode <= 299 {
		return nil
	}
	se := &errors.StatusError{}
	if err := c.decoder(res, se); err != nil {
		return err
	}
	return se
}

// DecodeResponse decodes the body of res into target. If there is no body, target is unchanged.
func (c *Client) DecodeResponse(res *http.Response, v interface{}) error {
	return c.decoder(res, v)
}
