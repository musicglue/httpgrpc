package server

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"

	"golang.org/x/net/context"

	"github.com/musicglue/httpgrpc"
	"github.com/musicglue/httpgrpc/utils"

	log "github.com/sirupsen/logrus"
)

// Server implements HTTPServer.  HTTPServer is a generated interface that gRPC
// servers must implement.
type Server struct {
	handler http.Handler
}

// New makes a new Server.
func New(handler http.Handler) *Server {
	return &Server{
		handler: handler,
	}
}

// Handle implements HTTPServer.
func (s Server) Handle(ctx context.Context, r *httpgrpc.HTTPRequest) (*httpgrpc.HTTPResponse, error) {
	req, err := http.NewRequest(r.Method, r.Url, ioutil.NopCloser(bytes.NewReader(r.Body)))
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	utils.ToHeader(r.Headers, req.Header)
	req.RequestURI = r.Url

	for k, v := range req.Header {
		log.Infof("received header: %s -> %s", k, v)
	}

	recorder := httptest.NewRecorder()
	s.handler.ServeHTTP(recorder, req)

	resp := &httpgrpc.HTTPResponse{
		Code:    int32(recorder.Code),
		Headers: utils.FromHeader(recorder.Header()),
		Body:    recorder.Body.Bytes(),
	}

	if recorder.Code/100 == 5 {
		return nil, httpgrpc.ErrorFromHTTPResponse(resp)
	}

	return resp, err
}
