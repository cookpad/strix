package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type strixUser struct {
	UserID    string
	Timestamp time.Time
}

type sessionManager struct {
	jwtSecret []byte
}

func newSessionManager(jwtSecret string) *sessionManager {
	mgr := &sessionManager{}

	if jwtSecret == "" {
		logger.Warn("jwt-secret is not set, then automatically generated")
		mgr.jwtSecret = []byte(genRandomSecret())
	} else {
		mgr.jwtSecret = []byte(jwtSecret)
	}

	return mgr
}

func (x *sessionManager) sign(user strixUser, c *gin.Context) error {
	ssn := sessions.Default(c)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user":      user.UserID,
		"timestamp": user.Timestamp,
	})
	signed, err := token.SignedString(x.jwtSecret)
	if err != nil {
		return errors.Wrapf(err, "Fail to sign JWT token: %v", token)
	}

	ssn.Set("jwt", signed)
	if err := ssn.Save(); err != nil {
		logger.WithError(err).Errorf("Fail to save cookie")
		return errors.Wrap(err, "Fail to save cookie")
	}

	return nil
}

func (x *sessionManager) validate(c *gin.Context) (*strixUser, error) {
	ssn := sessions.Default(c)
	cookie := ssn.Get("jwt")
	if cookie == nil {
		return nil, fmt.Errorf("No cookie")
	}

	raw, ok := cookie.(string)
	if !ok {
		return nil, fmt.Errorf("Invalid cookie data format: %v", cookie)
	}

	logger.WithField("jwt", raw).Info("Validating JWT token")

	token, err := jwt.Parse(raw, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return x.jwtSecret, nil
	})

	if token.Valid {
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return nil, errors.Wrapf(err, "Fail to get claims from Token: %v", claims)
		}

		logger.WithField("ts", claims["timestamp"].(string)).Info("timestamp")

		// Authentication success
		return &strixUser{
			UserID: claims["user"].(string),
		}, nil
	} else if ve, ok := err.(*jwt.ValidationError); ok {
		if ve.Errors&jwt.ValidationErrorMalformed != 0 {
			return nil, fmt.Errorf("That's not even a token")
		} else if ve.Errors&(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet) != 0 {
			return nil, fmt.Errorf("Timing is everything")
		} else {
			return nil, errors.Wrap(err, "Couldn't handle this token")
		}
	}

	return nil, errors.Wrap(err, "Couldn't handle this token:")
}

func setupAuth(mgr *sessionManager, r *gin.RouterGroup) error {
	r.GET("/", func(c *gin.Context) {
		user, err := mgr.validate(c)
		logger.WithField("user", user).Info("Auth")

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"msg": "Not authenticated"})
		} else {
			c.JSON(http.StatusOK, gin.H{"msg": "Authenticated"})
		}
	})

	return nil
}

func loadAuthGoogleConfig(configPath string) (*oauth2.Config, error) {
	raw, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	conf, err := google.ConfigFromJSON(raw, "https://www.googleapis.com/auth/userinfo.email")
	if err != nil {
		return nil, err
	}

	return conf, nil
}

func setupAuthGoogle(mgr *sessionManager, clientConfig string, r *gin.RouterGroup) error {
	conf, err := loadAuthGoogleConfig(clientConfig)
	if err != nil {
		return errors.Wrapf(err, "Fail to load Google OAuth client config: %s", clientConfig)
	}

	// Redirect to Google
	r.GET("/google", func(c *gin.Context) {
		url := conf.AuthCodeURL("state", oauth2.AccessTypeOnline)
		// fmt.Printf("Visit the URL for the auth dialog: %v", url)

		c.Redirect(302, url)
	})

	// Callback from Google
	r.GET("/google/callback", func(c *gin.Context) {
		if errmsg := c.Query("error"); errmsg != "" {
			c.String(http.StatusUnauthorized, "Auth error: "+errmsg)
			return
		}

		code := c.Query("code")
		if code == "" {
			c.String(400, "No auth code")
			return
		}

		ctx := context.Background()
		token, err := conf.Exchange(ctx, code)
		if err != nil {
			logger.WithError(err).Errorf("Fail to parse token from Google: %v", code)
			c.String(http.StatusInternalServerError, "Invalid Token, see system logs")
			return
		}

		client := conf.Client(ctx, token)
		resp, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")

		if err != nil {
			logger.WithError(err).Errorf("Fail to get user info from Google")
			c.String(http.StatusInternalServerError, "Fail to authentication, see system logs")
			return
		}

		defer resp.Body.Close()

		var googleUser struct {
			Sub           string `json:"sub"`
			Picture       string `json:""`
			Email         string `json:"email"`
			EmailVerified bool   `json:"email_verified"`
			HD            string `json:"hd"`
		}

		raw, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			logger.WithError(err).Errorf("Fail to read user info from Google")
			c.String(http.StatusInternalServerError, "Fail to authentication, see system logs")
			return
		}

		if err := json.Unmarshal(raw, &googleUser); err != nil {
			logger.WithError(err).WithField("raw", string(raw)).Errorf("Fail to parse user info from Google")
			c.String(http.StatusInternalServerError, "Fail to authentication, see system logs")
			return
		}

		logger.WithField("user", googleUser).Info("Got user info from Google")
		if !googleUser.EmailVerified {
			c.String(http.StatusUnauthorized, "Email is not verified: "+googleUser.Email)
			return
		}

		user := strixUser{
			UserID:    googleUser.Email,
			Timestamp: time.Now(),
		}
		if err := mgr.sign(user, c); err != nil {
			c.String(http.StatusInternalServerError, "Authentication procedure failed")
		}

		c.Redirect(302, "/")
	})

	return nil
}

func authGoogle(c *gin.Context) {
	c.String(http.StatusOK, "ok")
}

func authGoogleCallback(c *gin.Context) {
	c.String(http.StatusOK, "ok")
}
