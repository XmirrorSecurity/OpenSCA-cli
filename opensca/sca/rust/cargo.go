package rust

import (
	"io"
	"strings"

	"github.com/xmirrorsecurity/opensca-cli/opensca/logs"
	"github.com/xmirrorsecurity/opensca-cli/opensca/model"

	"github.com/BurntSushi/toml"
)

func ParseCargoLock(file *model.File) *model.DepGraph {

	cargo := struct {
		Pkgs []struct {
			Name         string   `toml:"name"`
			Version      string   `toml:"version"`
			dependencies []string `toml:"dependencies"`
		} `toml:"package"`
	}{}

	file.OpenReader(func(reader io.Reader) {
		_, err := toml.NewDecoder(reader).Decode(&cargo)
		if err != nil {
			logs.Warnf("parse %s fail:%s", file.Relpath, err)
		}
	})

	// 记录组件信息
	depMap := map[string]*model.DepGraph{}
	for _, c := range cargo.Pkgs {
		depMap[c.Name] = &model.DepGraph{
			Name:    c.Name,
			Version: c.Version,
		}
	}

	// 记录依赖关系
	for _, c := range cargo.Pkgs {
		for _, dependency := range c.dependencies {
			name := dependency
			i := strings.Index(dependency, " ")
			if i != -1 {
				name = dependency[:i]
			}
			depMap[c.Name].AppendChild(depMap[name])
		}
	}

	root := &model.DepGraph{Path: file.Relpath}
	for _, dep := range depMap {
		if len(dep.Parents) == 0 {
			root.AppendChild(dep)
		}
	}
	return root
}
