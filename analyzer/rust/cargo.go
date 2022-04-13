package rust

import (
	"sort"
	"strings"
	"util/logs"
	"util/model"

	"github.com/BurntSushi/toml"
)

type cargoPkg struct {
	Name         string   `toml:"name"`
	Version      string   `toml:"version"`
	DepStr       []string `toml:"dependencies"`
	Dependencies []struct {
		Name    string
		Version string
	} `toml:"-"`
}

func parseCargoLock(dirRoot *model.DirTree, depRoot *model.DepTree, file *model.FileData) []*model.DepTree {
	cargo := struct {
		Pkgs []*cargoPkg `toml:"package"`
	}{}
	cdepMap := map[string]*cargoPkg{}
	depMap := map[string]*model.DepTree{}
	directMap := map[string]*model.DepTree{}
	if err := toml.Unmarshal(file.Data, &cargo); err != nil {
		logs.Warn(err)
	}
	for _, pkg := range cargo.Pkgs {
		dep := model.NewDepTree(nil)
		dep.Name = pkg.Name
		dep.Version = model.NewVersion(pkg.Version)
		pkg.Dependencies = make([]struct {
			Name    string
			Version string
		}, len(pkg.DepStr))
		for i, str := range pkg.DepStr {
			name, version := str, ""
			index := strings.Index(str, " ")
			if index > -1 {
				name, version = str[:index], str[index+1:]
			}
			pkg.Dependencies[i] = struct {
				Name    string
				Version string
			}{Name: name, Version: version}
		}
		depMap[dep.Name] = dep
		directMap[dep.Name] = dep
		cdepMap[dep.Name] = pkg
	}
	// 找出未被依赖的作为直接依赖
	for _, pkg := range cargo.Pkgs {
		for _, d := range pkg.Dependencies {
			delete(directMap, d.Name)
		}
	}
	directDeps := []*model.DepTree{}
	for _, v := range directMap {
		directDeps = append(directDeps, v)
	}
	sort.Slice(directDeps, func(i, j int) bool {
		return directDeps[i].Name < directDeps[j].Name
	})
	for _, d := range directDeps {
		d.Parent = depRoot
		depRoot.Children = append(depRoot.Children, d)
	}
	// 从顶层开始构建
	q := make([]*model.DepTree, len(directDeps))
	copy(q, directDeps)
	exist := map[string]struct{}{}
	for len(q) > 0 {
		n := q[0]
		exist[n.Name] = struct{}{}
		if cdep, ok := cdepMap[n.Name]; ok {
			for _, d := range cdep.Dependencies {
				if _, ok := exist[d.Name]; !ok {
					exist[d.Name] = struct{}{}
					if sub, ok := depMap[d.Name]; ok {
						sub.Parent = n
						n.Children = append(n.Children, sub)
					}
				}
			}
		}
		q = append(q[1:], n.Children...)
	}
	return directDeps
}
