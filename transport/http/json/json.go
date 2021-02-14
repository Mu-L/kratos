package json

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

var (
	// MarshalOptions is a configurable JSON format marshaler.
	MarshalOptions = protojson.MarshalOptions{
		EmitUnpopulated: true,
	}
	// UnmarshalOptions is a configurable JSON format parser.
	UnmarshalOptions = protojson.UnmarshalOptions{
		DiscardUnknown: true,
	}
)

// DecodeRequest decode server reqeust from body.
func DecodeRequest(req *http.Request, v interface{}) error {
	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return err
	}
	defer req.Body.Close()
	if v, ok := v.(proto.Message); ok {
		return UnmarshalOptions.Unmarshal(data, v)
	}
	return json.Unmarshal(data, v)
}

// EncodeResponse encode server response to writer.
func EncodeResponse(res http.ResponseWriter, req *http.Request, v interface{}) error {
	var (
		err  error
		data []byte
	)
	if v, ok := v.(proto.Message); ok {
		if data, err = MarshalOptions.Marshal(v); err != nil {
			return err
		}
	} else {
		if data, err = json.Marshal(v); err != nil {
			return err
		}
	}
	res.Header().Set("content-type", "application/json")
	res.Write(data)
	return nil
}

// DecodeResponse decode client response from body.
func DecodeResponse(res *http.Response, v interface{}) error {
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if v, ok := v.(proto.Message); ok {
		return UnmarshalOptions.Unmarshal(data, v)
	}
	return json.Unmarshal(data, v)
}
