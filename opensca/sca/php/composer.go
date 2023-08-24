package php

import (
	"encoding/json"
	"io"

	"github.com/xmirrorsecurity/opensca-cli/opensca/logs"
	"github.com/xmirrorsecurity/opensca-cli/opensca/model"
)

type ComposerJson struct {
	Name       string            `json:"name"`
	License    string            `json:"license"`
	Require    map[string]string `json:"require"`
	RequireDev map[string]string `json:"require-dev"`
}

type ComposerLock struct {
	Pkgs []struct {
		Name     string            `json:"name"`
		Version  string            `json:"version"`
		Require  map[string]string `json:"require"`
		HomePage string            `json:"homepage"`
		Source   map[string]string `json:"source"`
	} `json:"packages"`
}

type ComposerRepo struct {
	Pkgs map[string][]struct {
		Version string            `json:"version"`
		Require map[string]string `json:"require"`
	} `json:"packages"`
}

func ParseComposerJson(file *model.File) *model.DepGraph {
	composer := ComposerJson{}
	file.OpenReader(func(reader io.Reader) {
		if err := json.NewDecoder(reader).Decode(&composer); err != nil {
			logs.Warn(err)
		}
	})
	root := &model.DepGraph{Name: composer.Name}
	root.AppendLicense(composer.License)
	return nil
}
