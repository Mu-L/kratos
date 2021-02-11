package http

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"google.golang.org/protobuf/proto"
)

var (
	// MIMEJSON is json content type.
	MIMEJSON = "application/json"
	// MIMEPROTOBUF is protobuf content type.
	MIMEPROTOBUF = "application/proto"
)

func stripContentType(contentType string) string {
	i := strings.Index(contentType, ";")
	if i != -1 {
		contentType = contentType[:i]
	}
	return contentType
}

func marshalForAccepts(req *http.Request, v interface{}) (string, []byte, error) {
	contentType := stripContentType(req.Header.Get("accept"))
	switch contentType {
	case MIMEPROTOBUF:
		data, err := proto.Marshal(v.(proto.Message))
		if err != nil {
			return "", nil, err
		}
		return MIMEPROTOBUF, data, nil
	default:
		data, err := json.Marshal(v)
		if err != nil {
			return "", nil, err
		}
		return MIMEJSON, data, nil
	}
}

// DefaultRequestDecoder default request decoder.
func DefaultRequestDecoder(req *http.Request, v interface{}) error {
	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return err
	}
	defer req.Body.Close()
	contentType := req.Header.Get("content-type")
	switch contentType {
	case MIMEJSON:
		if err = json.Unmarshal(data, v); err != nil {
			return err
		}
	case MIMEPROTOBUF:
		if err = proto.Unmarshal(data, v.(proto.Message)); err != nil {
			return err
		}
	default:
		if err := BindQuery(req, v.(proto.Message)); err != nil {
			return err
		}
	}
	return nil
}

// DefaultResponseEncoder is default response encoder.
func DefaultResponseEncoder(res http.ResponseWriter, req *http.Request, v interface{}) error {
	contentType, data, err := marshalForAccepts(req, v)
	if err != nil {
		return err
	}
	res.Header().Set("content-type", contentType)
	res.Write(data)
	return nil
}

// DefaultErrorEncoder is default errors encoder.
func DefaultErrorEncoder(res http.ResponseWriter, req *http.Request, err error) {
	code, se := StatusError(err)
	contentType, data, err := marshalForAccepts(req, se)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	res.Header().Set("content-type", contentType)
	res.WriteHeader(code)
	res.Write(data)
}
