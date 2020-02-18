package main

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type roundTripper func(*http.Request) (*http.Response, error)

func (f roundTripper) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

func reverseProxy(authz *authzService, apiKey, target string) (gin.HandlerFunc, error) {
	logger.WithFields(logrus.Fields{
		"target": target,
		"apikey": apiKey[:4] + "...",
		"authz":  authz,
	}).Info("build proxy")

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

		userID := userData.(string)
		user := authz.lookup(userID)
		if user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"msg": "Unauthorized user", "user": user})
			return
		}

		permittedTags := strings.Join(user.permitted(), ",")
		if permittedTags == "" {
			permittedTags = "*"
		}

		reqID := uuid.New().String()
		logger.WithFields(logrus.Fields{
			"user":          user.UserID,
			"permittedTags": permittedTags,
			"path":          c.FullPath(),
			"ipaddr":        c.ClientIP(),
			"user_agent":    c.Request.UserAgent(),
			"request_id":    reqID,
		}).Info("Audit log")

		(&httputil.ReverseProxy{
			Transport: roundTripper(requestHandler),
			Director: func(req *http.Request) {
				req.URL.Host = url.Host
				req.URL.Scheme = url.Scheme
				req.URL.Path = url.Path + req.URL.Path
				req.Header.Set("x-api-key", apiKey)
				req.Header.Set("x-permitted-tags", permittedTags)
				req.Header.Set("x-request-id", reqID)
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
	r.GET("/search/:search_id", proxy)
	r.GET("/search/:search_id/logs", proxy)
	r.GET("/search/:search_id/timeseries", proxy)

	return nil
}
