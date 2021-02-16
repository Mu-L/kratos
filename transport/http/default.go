package http

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/go-kratos/kratos/v2/encoding"
)

const baseContentType = "application"

func contentSubtype(contentType string) string {
	if contentType == baseContentType {
		return ""
	}
	if !strings.HasPrefix(contentType, baseContentType) {
		return ""
	}
	// guaranteed since != baseContentType and has baseContentType prefix
	switch contentType[len(baseContentType)] {
	case '/', ';':
		// this will return true for "application/grpc+" or "application/grpc;"
		// which the previous validContentType function tested to be valid, so we
		// just say that no content-subtype is specified in this case
		return contentType[len(baseContentType)+1:]
	default:
		return ""
	}
}

func defaultRequestDecoder(req *http.Request, v interface{}) error {
	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return err
	}
	defer req.Body.Close()
	subtype := contentSubtype(req.Header.Get("content-type"))
	codec := encoding.GetCodec(subtype)
	if codec == nil {
		return fmt.Errorf("unknown content-type error: %s", subtype)
	}
	return codec.Unmarshal(data, v)
}

func defaultResponseEncoder(res http.ResponseWriter, req *http.Request, v interface{}) error {
	subtype := contentSubtype(req.Header.Get("accept"))
	codec := encoding.GetCodec(subtype)
	if codec == nil {
		codec = encoding.GetCodec("json")
	}
	data, err := codec.Marshal(v)
	if err != nil {
		return err
	}
	res.Write(data)
	return nil
}

func defaultErrorEncoder(res http.ResponseWriter, req *http.Request, err error) {
	code, se := StatusError(err)
	subtype := contentSubtype(req.Header.Get("accept"))
	codec := encoding.GetCodec(subtype)
	if codec == nil {
		codec = encoding.GetCodec("json")
	}
	data, err := codec.Marshal(se)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	res.WriteHeader(code)
	res.Write(data)
}
