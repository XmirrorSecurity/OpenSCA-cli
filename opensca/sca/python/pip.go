package python

import (
	"encoding/json"
	"io"
	"strings"

	"github.com/xmirrorsecurity/opensca-cli/opensca/logs"
	"github.com/xmirrorsecurity/opensca-cli/opensca/model"
)

func ParsePipfile(file *model.File) *model.DepGraph {

	pip := struct {
		DevPackages map[string]string `toml:"dev-packages"`
		Packages    map[string]string `toml:"packages"`
	}{}

	root := &model.DepGraph{Path: file.Path()}

	_dep := model.NewDepGraphMap(nil, func(s ...string) *model.DepGraph { return &model.DepGraph{Name: s[0], Version: s[1]} }).LoadOrStore

	file.OpenReader(func(reader io.Reader) {
		if err := json.NewDecoder(reader).Decode(&pip); err != nil {
			logs.Warnf("unmarshal file %s err: %s", file.Path(), err)
		}
	})

	for name, version := range pip.Packages {
		root.AppendChild(_dep(name, version))
	}
	for name, version := range pip.DevPackages {
		dep := _dep(name, version)
		if dep == nil {
			continue
		}
		dep.Develop = true
		root.AppendChild(dep)
	}

	return root
}

func ParsePipfileLock(file *model.File) *model.DepGraph {

	lock := struct {
		Default map[string]struct {
			Version string `json:"version"`
		} `json:"default"`
	}{}

	root := &model.DepGraph{Path: file.Path()}

	_dep := model.NewDepGraphMap(nil, func(s ...string) *model.DepGraph { return &model.DepGraph{Name: s[0], Version: s[1]} }).LoadOrStore

	file.OpenReader(func(reader io.Reader) {
		if err := json.NewDecoder(reader).Decode(&lock); err != nil {
			logs.Warnf("unmarshal file %s err: %s", file.Path(), err)
		}
	})

	for name, v := range lock.Default {
		version := strings.TrimPrefix(v.Version, "==")
		root.AppendChild(_dep(name, version))
	}

	return root
}
