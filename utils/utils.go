package utils

import (
	"net/http"

	"github.com/musicglue/httpgrpc"
)

// ToHeader turns a grpc header to a http header
func ToHeader(hs []*httpgrpc.Header, header http.Header) {
	for _, h := range hs {
		header[h.Key] = h.Values
	}
}

// FromHeader turns a http header to a grpc header
func FromHeader(hs http.Header) []*httpgrpc.Header {
	result := make([]*httpgrpc.Header, 0, len(hs))
	for k, vs := range hs {
		result = append(result, &httpgrpc.Header{
			Key:    k,
			Values: vs,
		})
	}
	return result
}
