package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/urfave/cli"
)

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
		{
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
			Name: "hello-reply, r", Value: time.Now().String(),
			Usage:       "Reply message for /hello/revision",
			Destination: &args.HelloReply,
		},

		cli.StringFlag{
			Name:        "google-oauth-config, g",
			Usage:       "Google OAuth config JSON file",
			Destination: &args.GoogleOAuthConfig,
		},
		cli.StringFlag{
			Name:        "google-oauth-config-data, d",
			Usage:       "Google OAuth JSON config encoded as base64",
			EnvVar:      "GOOGLE_OAUTH",
			Destination: &args.GoogleOAuthConfigData,
		},
		cli.StringFlag{
			Name:        "jwt-secret, j",
			Usage:       "JWT secret to sign and validate token",
			EnvVar:      "JWT_SECRET",
			Destination: &args.JWTSecret,
		},
		cli.StringFlag{
			Name:        "api-key, k",
			Usage:       "API Key of Minerva",
			EnvVar:      "API_KEY",
			Destination: &args.APIKey,
		},
		cli.StringFlag{
			Name:        "authz-path, z",
			Usage:       "Authorization list json file path",
			Destination: &args.AuthzFilePath,
		},
	}
	app.ArgsUsage = "[endpoint]"

	app.Action = func(c *cli.Context) error {
		if c.NArg() != 1 {
			return fmt.Errorf("endpoint is required")
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
