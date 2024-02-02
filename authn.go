package main

import (
	"context"
	"encoding/base64"
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

const (
	cookieKey     = "jwt"
	tokenDuration = time.Hour * 24
)

type strixUser struct {
	UserID    string    `json:"user"`
	Image     string    `json:"image"`
	ExpiresAt time.Time `json:"expires_at"`
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
		"user":       user.UserID,
		"expires_at": user.ExpiresAt,
		"image":      user.Image,
	})
	signed, err := token.SignedString(x.jwtSecret)
	if err != nil {
		return errors.Wrapf(err, "fail to sign JWT token: %v", token)
	}

	ssn.Set(cookieKey, signed)
	if err := ssn.Save(); err != nil {
		logger.WithError(err).Errorf("fail to save cookie")
		return errors.Wrap(err, "fail to save cookie")
	}

	return nil
}

func claimToStrixUser(claims jwt.MapClaims) (*strixUser, error) {
	var user strixUser

	if v, ok := claims["user"].(string); ok {
		user.UserID = v
	} else {
		return nil, fmt.Errorf("missing 'user' field in token")
	}

	if v, ok := claims["image"].(string); ok {
		user.Image = v
	} else {
		return nil, fmt.Errorf("missing 'image' field in token")
	}

	if v, ok := claims["expires_at"].(string); ok {
		if expires, err := time.Parse("2006-01-02T15:04:05.999999Z07:00", v); err == nil {
			user.ExpiresAt = expires
		} else {
			return nil, errors.Wrapf(err, "fail to parse 'expires_at' field properly: %s", v)
		}
	} else {
		return nil, fmt.Errorf("missing 'expires_at' field in token")
	}

	return &user, nil
}

func (x *sessionManager) logout(c *gin.Context) {
	ssn := sessions.Default(c)
	ssn.Delete(cookieKey)
	if err := ssn.Save(); err != nil {
		logger.WithError(err).Errorf("Fail to delete cookie, but nothing to do")
	}
}

func (x *sessionManager) validate(c *gin.Context) (*strixUser, error) {
	ssn := sessions.Default(c)
	cookie := ssn.Get(cookieKey)
	if cookie == nil {
		return nil, fmt.Errorf("no cookie")
	}

	raw, ok := cookie.(string)
	if !ok {
		return nil, fmt.Errorf("invalid cookie data format: %v", cookie)
	}

	token, err := jwt.Parse(raw, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return x.jwtSecret, nil
	})

	if token.Valid {
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return nil, errors.Wrapf(err, "Fail to get claims from Token: %v", claims)
		}

		user, err := claimToStrixUser(claims)
		if err != nil {
			return nil, err
		}

		if time.Now().After(user.ExpiresAt) {
			return nil, fmt.Errorf("Token is already expired: %s", user.ExpiresAt)
		}

		return user, nil
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
			logger.WithError(err).Info("Authentication fail")
			c.JSON(http.StatusUnauthorized, gin.H{"msg": "Not authenticated"})
		} else {
			c.JSON(http.StatusOK, gin.H{"msg": "Authenticated", "user": user})
		}
	})

	r.GET("/logout", func(c *gin.Context) {
		mgr.logout(c)
		c.Redirect(http.StatusFound, "/")
	})

	return nil
}

func setupAuthGoogleConfigFile(mgr *sessionManager, configPath string, r *gin.RouterGroup) error {
	raw, err := ioutil.ReadFile(configPath)
	if err != nil {
		return err
	}

	conf, err := google.ConfigFromJSON(raw, "https://www.googleapis.com/auth/userinfo.email")
	if err != nil {
		return err
	}

	return setupAuthGoogle(mgr, conf, r)
}

func setupAuthGoogleBase64(mgr *sessionManager, configData string, r *gin.RouterGroup) error {
	raw, err := base64.StdEncoding.DecodeString(configData)
	if err != nil {
		return errors.Wrapf(err, "Fail to decode Google OAuth data: %s", configData)
	}

	conf, err := google.ConfigFromJSON(raw, "https://www.googleapis.com/auth/userinfo.email")
	if err != nil {
		return errors.Wrap(err, "Fail to load JSON config")
	}

	return setupAuthGoogle(mgr, conf, r)
}

func setupAuthGoogle(mgr *sessionManager, conf *oauth2.Config, r *gin.RouterGroup) error {
	// Redirect to Google
	r.GET("/google", func(c *gin.Context) {
		url := conf.AuthCodeURL("state", oauth2.AccessTypeOnline)
		// fmt.Printf("Visit the URL for the auth dialog: %v", url)

		c.Redirect(http.StatusFound, url)
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
			Picture       string `json:"picture"`
			Email         string `json:"email"`
			EmailVerified bool   `json:"email_verified"`
			HD            string `json:"hd"`
		}

		raw, err := ioutil.ReadAll(resp.Body)
		logger.WithField("user", string(raw)).Info("userinfo")
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
			Image:     googleUser.Picture,
			ExpiresAt: time.Now().Add(tokenDuration),
		}
		if err := mgr.sign(user, c); err != nil {
			c.String(http.StatusInternalServerError, "Authentication procedure failed")
		}

		c.Redirect(http.StatusFound, "/")
	})

	return nil
}

func authGoogle(c *gin.Context) {
	c.String(http.StatusOK, "ok")
}

func authGoogleCallback(c *gin.Context) {
	c.String(http.StatusOK, "ok")
}
