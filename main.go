package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
)

var logger = logrus.New()

var logLevelMap = map[string]logrus.Level{
	"trace": logrus.TraceLevel,
	"debug": logrus.DebugLevel,
	"info":  logrus.InfoLevel,
	"warn":  logrus.WarnLevel,
	"error": logrus.ErrorLevel,
}

type arguments struct {
	LogLevel       string
	Endpoint       string
	BindAddress    string
	BindPort       int
	StaticContents string

	// Google OAuth options
	GoogleOAuthConfig string

	// JWT
	JWTSecret string
}

func runServer(args arguments) error {
	level, ok := logLevelMap[args.LogLevel]
	if !ok {
		return fmt.Errorf("Invalid log level: %s", args.LogLevel)
	}
	logger.SetLevel(level)
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.WithFields(logrus.Fields{
		"args": args,
	}).Info("Given options")

	helloReply := os.Getenv("HELLO_REPLY")
	if helloReply == "" {
		helloReply = time.Now().String()
	}

	r := gin.Default()
	store := cookie.NewStore([]byte("auth"))
	r.Use(sessions.Sessions("strix", store))
	r.Use(static.Serve("/", static.LocalFile(args.StaticContents, false)))
	/*
		r.LoadHTMLGlob(path.Join(args.TemplatePath, "*"))
		r.GET("/", func(c *gin.Context) {
			c.HTML(http.StatusOK, "index.html", gin.H{
				"googleOAuth": (args.GoogleOAuthConfig != ""),
			})
		})
	*/
	r.GET("/hello/revision", func(c *gin.Context) {
		c.String(200, helloReply)
	})

	ssnMgr := newSessionManager(args.JWTSecret)

	authGroup := r.Group("/auth")
	if err := setupAuth(ssnMgr, authGroup); err != nil {
		return err
	}
	if args.GoogleOAuthConfig != "" {
		if err := setupAuthGoogle(ssnMgr, args.GoogleOAuthConfig, authGroup); err != nil {
			return err
		}
	}

	apiGroup := r.Group("/api/v1")
	if err := setupAPI(ssnMgr, args.Endpoint, apiGroup); err != nil {
		return err
	}

	if err := r.Run(fmt.Sprintf("%s:%d", args.BindAddress, args.BindPort)); err != nil {
		return err
	}

	return nil
}

func genRandomSecret() string {
	const randomSecretLength = 32
	letters := "abcdefghijklmnopqrstuvwxyz" +
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"0123456789"

	src := rand.NewSource(time.Now().UnixNano())
	rnd := rand.New(src)

	var secret string
	for i := 0; i < randomSecretLength; i++ {
		n := rnd.Intn(len(letters))
		secret = secret + string(letters[n])
	}

	return secret
}

func main() {
	var args arguments

	app := cli.NewApp()
	app.Name = "strix"
	app.Usage = "Web UI for Minerva (security log search system)"
	app.Authors = []cli.Author{
		cli.Author{
			Name:  "Masayoshi Mizutani",
			Email: "mizutani@cookpad.com",
		},
	}

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name: "log-level, l", Value: "info",
			Usage:       "Log level [trace,debug,info,warn,error]",
			Destination: &args.LogLevel,
		},
		cli.StringFlag{
			Name: "addr, a", Value: "127.0.0.1",
			Usage:       "Bind address",
			Destination: &args.BindAddress,
		},
		cli.IntFlag{
			Name: "port, p", Value: 9080,
			Usage:       "Bind port",
			Destination: &args.BindPort,
		},
		cli.StringFlag{
			Name: "static, s", Value: "./static",
			Usage:       "Static contents path",
			Destination: &args.StaticContents,
		},

		cli.StringFlag{
			Name:        "google-oauth-config, g",
			Usage:       "Google OAuth config JSON file",
			Destination: &args.GoogleOAuthConfig,
		},
		cli.StringFlag{
			Name:        "jwt-secret, j",
			Value:       genRandomSecret(),
			Usage:       "JWT secret to sign and validate token",
			Destination: &args.GoogleOAuthConfig,
		},
	}
	app.ArgsUsage = "[endpoint]"

	app.Action = func(c *cli.Context) error {
		if c.NArg() != 1 {
			return fmt.Errorf("Endpoint is required")
		}
		args.Endpoint = c.Args().Get(0)

		if err := runServer(args); err != nil {
			return err
		}
		return nil
	}

	if err := app.Run(os.Args); err != nil {
		logger.WithError(err).Fatal("Fatal Error")
	}
}
