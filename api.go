package main

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type roundTripper func(*http.Request) (*http.Response, error)

func (f roundTripper) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

func reverseProxy(target string) (gin.HandlerFunc, error) {
	logger.WithField("target", target).Info("proxy")
	url, err := url.Parse(target)
	if err != nil {
		return nil, errors.Wrapf(err, "Fail to parse endpoint URL: %v", target)
	}

	requestHandler := func(req *http.Request) (*http.Response, error) {
		req.Host = url.Host
		return http.DefaultTransport.RoundTrip(req)
	}

	proxy := &httputil.ReverseProxy{
		Transport: roundTripper(requestHandler),
		Director: func(req *http.Request) {
			req.URL.Host = url.Host
			req.URL.Scheme = url.Scheme
			req.URL.Path = url.Path + req.URL.Path
		},
	}

	return func(c *gin.Context) {
		proxy.ServeHTTP(c.Writer, c.Request)
	}, nil
}

func setupAPI(ssnMgr *sessionManager, endpoint string, r *gin.RouterGroup) error {
	proxy, err := reverseProxy(endpoint)
	if err != nil {
		return err
	}

	r.POST("/search", proxy)
	r.GET("/search/:search_id/logs", proxy)
	r.GET("/search/:search_id/timeseries", proxy)

	return nil
}
