package rust

import (
	"io"
	"strings"

	"github.com/xmirrorsecurity/opensca-cli/v3/opensca/logs"
	"github.com/xmirrorsecurity/opensca-cli/v3/opensca/model"

	"github.com/BurntSushi/toml"
)

// ParseCargoLock 解析Cargo.lock文件
func ParseCargoLock(file *model.File) *model.DepGraph {

	cargo := struct {
		Pkgs []struct {
			Name         string   `toml:"name"`
			Version      string   `toml:"version"`
			Dependencies []string `toml:"dependencies"`
		} `toml:"package"`
	}{}

	file.OpenReader(func(reader io.Reader) {
		_, err := toml.NewDecoder(reader).Decode(&cargo)
		if err != nil {
			logs.Warnf("parse %s fail:%s", file.Relpath(), err)
		}
	})

	// 记录组件信息
	depMap := map[string]*model.DepGraph{}
	_dep := model.NewDepGraphMap(nil, func(s ...string) *model.DepGraph {
		return &model.DepGraph{
			Name:    s[0],
			Version: s[1],
		}
	}).LoadOrStore
	for _, c := range cargo.Pkgs {
		depMap[c.Name] = _dep(c.Name, c.Version)
	}

	// 记录依赖关系
	for _, c := range cargo.Pkgs {
		dep := _dep(c.Name, c.Version)
		for _, dependency := range c.Dependencies {
			i := strings.Index(dependency, " ")
			if i != -1 {
				name, version := dependency[:i], dependency[i+1:]
				dep.AppendChild(_dep(name, version))
			} else {
				dep.AppendChild(depMap[dependency])
			}
		}
	}

	root := &model.DepGraph{Path: file.Relpath()}
	for _, dep := range depMap {
		if len(dep.Parents) == 0 {
			root.AppendChild(dep)
		}
	}

	return root
}
