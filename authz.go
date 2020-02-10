package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/pkg/errors"
)

type authzService struct {
	AllowTable map[string][]string `json:"allow"`
}

func newAuthzService(filePath string) (*authzService, error) {
	raw, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, errors.Wrapf(err, "Fail to load authz file: %s", filePath)
	}

	srv := authzService{
		AllowTable: make(map[string][]string),
	}
	if err := json.Unmarshal(raw, &srv); err != nil {
		return nil, errors.Wrapf(err, "Fail to parse authz file: %s", filePath)
	}

	logger.WithField("authz", srv.AllowTable).Info("Read authorization table")
	return &srv, nil
}

func (x *authzService) allowedTags(user string) ([]string, error) {
	allowed, ok := x.AllowTable[user]
	if !ok {
		return nil, fmt.Errorf("Not permitted")
	}

	return allowed, nil
}
