package client

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/musicglue/httpgrpc/resolver"

	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	"github.com/mwitkow/go-grpc-middleware"
	"github.com/opentracing/opentracing-go"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/musicglue/httpgrpc"
	"github.com/musicglue/httpgrpc/utils"
	"github.com/weaveworks/common/user"
)

// Client is a http.Handler that forwards the request over gRPC.
type Client struct {
	mtx       sync.RWMutex
	service   string
	namespace string
	port      string
	client    httpgrpc.HTTPClient
	conn      *grpc.ClientConn
}

func clientUserHeaderInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	ctx, err := user.InjectIntoGRPCRequest(ctx)
	if err != nil {
		return err
	}

	return invoker(ctx, method, req, reply, cc, opts...)
}

// New makes a new Client, given a kubernetes service address.
func New(address string) (*Client, error) {
	address, dialOptions, err := ParseURL(address)
	if err != nil {
		return nil, err
	}

	dialOptions = append(
		dialOptions,
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(grpc_middleware.ChainUnaryClient(
			otgrpc.OpenTracingClientInterceptor(opentracing.GlobalTracer()),
			clientUserHeaderInterceptor,
		)),
	)

	conn, err := grpc.Dial(address, dialOptions...)
	if err != nil {
		return nil, err
	}

	return &Client{
		client: httpgrpc.NewHTTPClient(conn),
		conn:   conn,
	}, nil
}

// ServeHTTP implements http.Handler
func (c *Client) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	req := &httpgrpc.HTTPRequest{
		Method:  r.Method,
		Url:     r.RequestURI,
		Body:    body,
		Headers: utils.FromHeader(r.Header),
	}

	resp, err := c.client.Handle(r.Context(), req)
	if err != nil {
		// Some errors will actually contain a valid resp, just need to unpack it
		var ok bool
		resp, ok = httpgrpc.HTTPResponseFromError(err)

		if !ok {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	utils.ToHeader(resp.Headers, w.Header())
	w.WriteHeader(int(resp.Code))
	if _, err := w.Write(resp.Body); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// ParseURL deals with direct:// style URLs, as well as kubernetes:// urls.
// For backwards compatibility it treats URLs without schems as kubernetes://.
func ParseURL(unparsed string) (string, []grpc.DialOption, error) {
	parsed, err := url.Parse(unparsed)
	if err != nil {
		return "", nil, err
	}

	scheme, host := parsed.Scheme, parsed.Host
	if !strings.Contains(unparsed, "://") {
		scheme, host = "kubernetes", unparsed
	}

	switch scheme {
	case "direct":
		return host, nil, err

	case "kubernetes":
		host, port, err := net.SplitHostPort(host)
		if err != nil {
			return "", nil, err
		}
		parts := strings.SplitN(host, ".", 2)
		service, namespace := parts[0], "default"
		if len(parts) == 2 {
			namespace = parts[1]
		}
		balancer := resolver.NewWithNamespace(namespace)
		address := fmt.Sprintf("kubernetes://%s:%s", service, port)
		dialOptions := []grpc.DialOption{balancer.DialOption()}
		return address, dialOptions, nil

	default:
		return "", nil, fmt.Errorf("unrecognised scheme: %s", parsed.Scheme)
	}
}
