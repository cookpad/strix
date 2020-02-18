package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"regexp"

	"github.com/pkg/errors"
)

type authzUser struct {
	UserID  string `json:"user_id"`
	Role    string `json:"role"`
	rolePtr *authzRole
}

func (x *authzUser) allowed() []string {
	return x.rolePtr.AllowedTags
}

type authzRole struct {
	Name        string   `json:"name"`
	AllowedTags []string `json:"allowed_tags"`
}

type authzRule struct {
	UserRegex string `json:"user_regex"`
	Role      string `json:"role"`
	regex     *regexp.Regexp
	rolePtr   *authzRole
}

type authzService struct {
	Users   []*authzUser `json:"users"`
	Roles   []*authzRole `json:"roles"`
	Rules   []*authzRule `json:"rules"`
	UserMap map[string]*authzUser
	RoleMap map[string]*authzRole
}

func newAuthzServiceFromFile(filePath string) (*authzService, error) {
	raw, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, errors.Wrapf(err, "Fail to load authz file: %s", filePath)
	}

	return newAuthzService(raw)
}

func newAuthzService(raw []byte) (*authzService, error) {
	var srv authzService

	if err := json.Unmarshal(raw, &srv); err != nil {
		return nil, errors.Wrapf(err, "Fail to parse authz data json: %s", string(raw))
	}

	srv.UserMap = map[string]*authzUser{}
	srv.RoleMap = map[string]*authzRole{}

	for _, r := range srv.Roles {
		if _, ok := srv.RoleMap[r.Name]; ok {
			return nil, fmt.Errorf("Role '%s' is duplicated", r.Name)
		}
		srv.RoleMap[r.Name] = r
	}

	for _, u := range srv.Users {
		if _, ok := srv.UserMap[u.UserID]; ok {
			return nil, fmt.Errorf("User '%s' is duplicated", u.UserID)
		}

		role, ok := srv.RoleMap[u.Role]
		if !ok {
			return nil, fmt.Errorf("Role '%s' of User '%s' is not found", u.Role, u.UserID)
		}
		u.rolePtr = role
		srv.UserMap[u.UserID] = u
	}

	for _, rule := range srv.Rules {
		role, ok := srv.RoleMap[rule.Role]
		if !ok {
			return nil, fmt.Errorf("Role '%s' of Rule '%s' is not found", rule.Role, rule.UserRegex)
		}
		rule.rolePtr = role

		ptn, err := regexp.Compile(rule.UserRegex)
		if err != nil {
			return nil, fmt.Errorf("Fail to compile regex of a rule: %s", rule.UserRegex)
		}
		rule.regex = ptn
	}

	logger.WithField("authz", srv).Info("Read authorization table")
	return &srv, nil
}

func (x *authzService) lookup(userID string) *authzUser {
	user, ok := x.UserMap[userID]
	if !ok {
		for _, rule := range x.Rules {
			if rule.regex.MatchString(userID) {
				newUser := &authzUser{UserID: userID, rolePtr: rule.rolePtr}
				x.UserMap[userID] = newUser
				return newUser
			}
		}
	}

	return user
}
