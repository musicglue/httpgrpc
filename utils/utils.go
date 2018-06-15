package utils

import (
	"net/http"

	"github.com/musicglue/httpgrpc"
)

var hopHeaders = []string{
	"Connection",
	"Content-Length",
	"Proxy-Connection", // non-standard but still sent by libcurl and rejected by e.g. google
	"Keep-Alive",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	"Te",      // canonicalized version of "TE"
	"Trailer", // not Trailers per URL above; http://www.rfc-editor.org/errata_search.php?eid=4522
	"Transfer-Encoding",
	"Upgrade",
}

var skipHopHeader = make(map[string]bool, len(hopHeaders))

func init() {
	for _, v := range hopHeaders {
		skipHopHeader[v] = true
	}
}

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
		if skipHopHeader[k] {
			continue
		}

		result = append(result, &httpgrpc.Header{
			Key:    k,
			Values: vs,
		})
	}
	return result
}
