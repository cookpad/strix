package main

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type roundTripper func(*http.Request) (*http.Response, error)

func (f roundTripper) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

func reverseProxy(authz *authzService, apiKey, target string) (gin.HandlerFunc, error) {
	logger.WithField("target", target).Info("proxy")
	url, err := url.Parse(target)
	if err != nil {
		return nil, errors.Wrapf(err, "Fail to parse endpoint URL: %v", target)
	}

	requestHandler := func(req *http.Request) (*http.Response, error) {
		req.Host = url.Host
		return http.DefaultTransport.RoundTrip(req)
	}

	return func(c *gin.Context) {
		userData, ok := c.Get("user")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"msg": "Unauthenticated request"})
			return
		}

		user := userData.(string)
		allowed, ok := authz.AllowTable[user]
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"msg": "Unauthorized user", "user": user})
			return
		}

		tags := strings.Join(allowed, ",")
		if tags == "" {
			tags = "*"
		}

		(&httputil.ReverseProxy{
			Transport: roundTripper(requestHandler),
			Director: func(req *http.Request) {
				req.URL.Host = url.Host
				req.URL.Scheme = url.Scheme
				req.URL.Path = url.Path + req.URL.Path
				req.Header.Set("x-api-key", apiKey)
				req.Header.Set("minerva-allowed-tags", tags)
			},
		}).ServeHTTP(c.Writer, c.Request)

	}, nil
}

func setupAPI(authz *authzService, apiKey, endpoint string, r *gin.RouterGroup) error {
	proxy, err := reverseProxy(authz, apiKey, endpoint)
	if err != nil {
		return err
	}

	r.POST("/search", proxy)
	r.GET("/search/:search_id/logs", proxy)
	r.GET("/search/:search_id/timeseries", proxy)

	return nil
}
