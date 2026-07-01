package proxy

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/kayodeayelegun/ai-gateway/internal/requestid"
	"github.com/kayodeayelegun/ai-gateway/pkg/response"
)

// New returns an http.Handler that reverse-proxies to upstreamBase, rewriting
// each outbound request path to upstreamPath (for example base
// http://localhost:9001 with path /chat).
func New(upstreamBase, upstreamPath string, logger *slog.Logger) (http.Handler, error) {
	target, err := url.Parse(upstreamBase)
	if err != nil {
		return nil, fmt.Errorf("parse upstream base %q: %w", upstreamBase, err)
	}
	if target.Scheme == "" || target.Host == "" {
		return nil, fmt.Errorf("upstream base %q must include scheme and host", upstreamBase)
	}

	baseTransport := http.DefaultTransport.(*http.Transport).Clone()
	proxy := &httputil.ReverseProxy{
		Rewrite: func(preq *httputil.ProxyRequest) {
			preq.Out.URL.Scheme = target.Scheme
			preq.Out.URL.Host = target.Host
			preq.Out.URL.Path = upstreamPath
			preq.Out.URL.RawPath = ""
			preq.Out.URL.RawQuery = preq.In.URL.RawQuery
			if id := requestid.FromContext(preq.In.Context()); id != "" {
				preq.Out.Header.Set(requestid.Header, id)
			}
		},
		Transport: &contextTransport{base: baseTransport},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			logger.Error("upstream request failed",
				"request_id", requestid.FromContext(r.Context()),
				"method", r.Method,
				"path", r.URL.Path,
				"upstream", upstreamBase+upstreamPath,
				"error", err,
			)
			response.Error(w, http.StatusBadGateway, "bad gateway")
		},
	}

	return proxy, nil
}

// contextTransport wraps a RoundTripper so outbound calls respect request
// context cancellation and deadlines from upstream middleware.
type contextTransport struct {
	base http.RoundTripper
}

func (t *contextTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if err := req.Context().Err(); err != nil {
		return nil, err
	}
	return t.base.RoundTrip(req)
}
