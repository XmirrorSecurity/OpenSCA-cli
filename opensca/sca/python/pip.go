package python

import (
	"encoding/json"
	"io"
	"strings"

	"github.com/xmirrorsecurity/opensca-cli/v3/opensca/logs"
	"github.com/xmirrorsecurity/opensca-cli/v3/opensca/model"
)

func ParsePipfile(file *model.File) *model.DepGraph {

	pip := struct {
		DevPackages map[string]string `toml:"dev-packages"`
		Packages    map[string]string `toml:"packages"`
	}{}

	root := &model.DepGraph{Path: file.Relpath()}

	_dep := model.NewDepGraphMap(nil, func(s ...string) *model.DepGraph { return &model.DepGraph{Name: s[0], Version: s[1]} }).LoadOrStore

	file.OpenReader(func(reader io.Reader) {
		if err := json.NewDecoder(reader).Decode(&pip); err != nil {
			logs.Warnf("unmarshal file %s err: %s", file.Relpath(), err)
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

	root := &model.DepGraph{Path: file.Relpath()}

	_dep := model.NewDepGraphMap(nil, func(s ...string) *model.DepGraph { return &model.DepGraph{Name: s[0], Version: s[1]} }).LoadOrStore

	file.OpenReader(func(reader io.Reader) {
		if err := json.NewDecoder(reader).Decode(&lock); err != nil {
			logs.Warnf("unmarshal file %s err: %s", file.Relpath(), err)
		}
	})

	for name, v := range lock.Default {
		version := strings.TrimPrefix(v.Version, "==")
		root.AppendChild(_dep(name, version))
	}

	return root
}

func ParseRequirementTxt(file *model.File) *model.DepGraph {

	root := &model.DepGraph{Path: file.Relpath()}

	file.ReadLine(func(line string) {

		if strings.HasPrefix(line, `-r`) {
			return
		}

		if i := strings.Index(line, "#"); i != -1 {
			line = line[:i]
		}

		line = strings.TrimSpace(strings.Split(line, ";")[0])
		if strings.ContainsAny(line, `#$%`) || len(line) == 0 {
			return
		}

		if i := strings.IndexAny(line, "=!<>~"); i == -1 {
			root.AppendChild(&model.DepGraph{Name: line})
		} else {
			root.AppendChild(&model.DepGraph{Name: line[:i], Version: strings.TrimPrefix(line[i:], "==")})
		}

	})

	return root
}

func ParseRequirementIn(file *model.File) *model.DepGraph {

	root := &model.DepGraph{Path: file.Relpath()}

	file.ReadLine(func(line string) {

		if strings.HasPrefix(line, `#`) {
			return
		}

		words := strings.Fields(line)
		if len(words) > 0 {
			dep := &model.DepGraph{}
			dep.Name = words[0]
			dep.Version = strings.Join(words[1:], "")
			root.AppendChild(dep)
		}

	})

	return root
}
