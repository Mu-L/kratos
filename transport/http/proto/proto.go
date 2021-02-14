package proto

import (
	"errors"
	"io/ioutil"
	"net/http"

	"google.golang.org/protobuf/proto"
)

// DecodeRequest decode server reqeust from body.
func DecodeRequest(req *http.Request, v interface{}) error {
	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return err
	}
	defer req.Body.Close()
	msg, ok := v.(proto.Message)
	if !ok {
		return errors.New("request does not implement proto.Message")
	}
	return proto.Unmarshal(data, msg)
}

// EncodeResponse encode server response to writer.
func EncodeResponse(res http.ResponseWriter, req *http.Request, v interface{}) error {
	msg, ok := v.(proto.Message)
	if !ok {
		return errors.New("response does not implement proto.Message")
	}
	data, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	res.Header().Set("content-type", "application/proto")
	res.Write(data)
	return nil
}

// DecodeResponse decode client response from body.
func DecodeResponse(res *http.Response, v interface{}) error {
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	msg, ok := v.(proto.Message)
	if !ok {
		return errors.New("response does not implement proto.Message")
	}
	return proto.Unmarshal(data, msg)
}
