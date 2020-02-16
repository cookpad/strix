package main

import (
	"fmt"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type arguments struct {
	LogLevel       string
	Endpoint       string
	BindAddress    string
	BindPort       int
	StaticContents string
	HelloReply     string
	APIKey         string
	AuthzFilePath  string

	// Google OAuth options
	GoogleOAuthConfig     string
	GoogleOAuthConfigData string

	// JWT
	JWTSecret string
}

func runServer(args arguments) error {
	if err := setupLogger(args.LogLevel); err != nil {
		return err
	}

	logger.WithFields(logrus.Fields{
		"args": args,
	}).Info("Given options")

	r := gin.Default()
	store := cookie.NewStore([]byte("auth"))
	r.Use(sessions.Sessions("strix", store))
	r.Use(static.Serve("/", static.LocalFile(args.StaticContents, false)))

	r.GET("/hello/revision", func(c *gin.Context) {
		c.String(200, args.HelloReply)
	})

	// Setup session manager
	authz, err := newAuthzService(args.AuthzFilePath)
	if err != nil {
		return err
	}

	ssnMgr := newSessionManager(args.JWTSecret)
	authCheck := func(c *gin.Context) {
		user, err := ssnMgr.validate(c)
		if err != nil {
			logger.WithError(err).Warn("Authentication Fail")
			c.JSON(http.StatusUnauthorized, gin.H{"msg": "Authentication failed"})
		} else {
			c.Set("user", user.UserID)
			c.Next()
		}
	}

	// Auth route group
	authGroup := r.Group("/auth")
	if err := setupAuth(ssnMgr, authGroup); err != nil {
		return err
	}
	if args.GoogleOAuthConfig != "" {
		if err := setupAuthGoogleConfigFile(ssnMgr, args.GoogleOAuthConfig, authGroup); err != nil {
			return err
		}
	}
	if args.GoogleOAuthConfigData != "" {
		if err := setupAuthGoogleBase64(ssnMgr, args.GoogleOAuthConfigData, authGroup); err != nil {
			return err
		}
	}

	// API route group
	apiGroup := r.Group("/api/v1")
	apiGroup.Use(authCheck)
	if err := setupAPI(authz, args.APIKey, args.Endpoint, apiGroup); err != nil {
		return err
	}

	// Start server
	if err := r.Run(fmt.Sprintf("%s:%d", args.BindAddress, args.BindPort)); err != nil {
		return err
	}

	return nil
}
