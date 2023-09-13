package rust

import (
	"io"
	"strings"

	"github.com/xmirrorsecurity/opensca-cli/opensca/logs"
	"github.com/xmirrorsecurity/opensca-cli/opensca/model"

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
	for _, c := range cargo.Pkgs {
		depMap[c.Name] = &model.DepGraph{
			Name:    c.Name,
			Version: c.Version,
		}
	}

	// 记录依赖关系
	for _, c := range cargo.Pkgs {
		for _, dependency := range c.Dependencies {
			name := dependency
			i := strings.Index(dependency, " ")
			if i != -1 {
				name = dependency[:i]
			}
			depMap[c.Name].AppendChild(depMap[name])
		}
	}

	for _, dep := range depMap {
		if len(dep.Parents) == 0 {
			dep.Path = file.Relpath()
			return dep
		}
	}

	return nil
}
