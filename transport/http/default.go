package http

import (
	"net/http"
	"strings"

	"github.com/go-kratos/kratos/v2/transport/http/json"
	"github.com/go-kratos/kratos/v2/transport/http/proto"
)

var (
	// DefaultRequestDecoders is default request encoders.
	DefaultRequestDecoders = map[string]DecodeRequestFunc{
		"application/json":  json.DecodeRequest,
		"application/proto": proto.DecodeRequest,
	}
	// DefaultResponseEncoders is default response encoders.
	DefaultResponseEncoders = map[string]EncodeResponseFunc{
		"application/json":  json.EncodeResponse,
		"application/proto": proto.EncodeResponse,
	}
	// DefaultResponseDecoders is default response decoders.
	DefaultResponseDecoders = map[string]DecodeResponseFunc{
		"application/json":  json.DecodeResponse,
		"application/proto": proto.DecodeResponse,
	}
)

func stripContentType(contentType string) string {
	idx := strings.Index(contentType, ";")
	if idx != -1 {
		contentType = contentType[:idx]
	}
	return contentType
}

// DefaultResponseDecoder is default response decoders.
func DefaultResponseDecoder(res *http.Response, v interface{}) error {
	contentType := stripContentType(res.Header.Get("content-type"))
	decode, ok := DefaultResponseDecoders[contentType]
	if ok {
		return decode(res, v)
	}
	return json.DecodeResponse(res, v)

}

// DefaultRequestDecoder is default request decoder.
func DefaultRequestDecoder(req *http.Request, v interface{}) error {
	contentType := stripContentType(req.Header.Get("content-type"))
	decode, ok := DefaultRequestDecoders[contentType]
	if ok {
		return decode(req, v)
	}
	return json.DecodeRequest(req, v)
}

// DefaultResponseEncoder is default response encoder.
func DefaultResponseEncoder(res http.ResponseWriter, req *http.Request, v interface{}) error {
	contentType := stripContentType(req.Header.Get("accept"))
	encode, ok := DefaultResponseEncoders[contentType]
	if ok {
		return encode(res, req, v)
	}
	return json.EncodeResponse(res, req, v)
}

// DefaultErrorEncoder is default errors encoder.
func DefaultErrorEncoder(res http.ResponseWriter, req *http.Request, err error) {
	code, se := StatusError(err)
	res.WriteHeader(code)
	encode, ok := DefaultResponseEncoders[stripContentType(req.Header.Get("accept"))]
	if !ok {
		encode = json.EncodeResponse
	}
	if err := encode(res, req, se); err != nil {
		res.WriteHeader(http.StatusInternalServerError)
	}
}
