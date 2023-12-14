package javascript

import (
	"strings"

	"github.com/xmirrorsecurity/opensca-cli/v3/opensca/logs"
	"github.com/xmirrorsecurity/opensca-cli/v3/opensca/model"
)

type YarnLock struct {
	Name         string
	Version      string
	Dependencies map[string]string
}

// ParseYarnLock 解析yarn.lock文件结构
func ParseYarnLock(file *model.File) map[string]*YarnLock {

	/*
		  name@version[, name@version]:
		    version "xxx"
			dependencies:
			  name "xxx"
			  name "xxx"
	*/

	lock := map[string]*YarnLock{}

	var lastDep *YarnLock

	file.ReadLine(func(line string) {

		if strings.HasPrefix(line, "#") {
			return
		}

		if !strings.HasPrefix(line, " ") && strings.HasSuffix(line, ":") {
			lastDep = &YarnLock{Dependencies: map[string]string{}}
			for _, tag := range strings.Split(line, ",") {
				i := strings.LastIndex(tag, "@")
				if i == -1 {
					logs.Warnf("parse file %s line: %s fail", file.Relpath(), line)
					continue
				}
				name := strings.Trim(tag[:i], ` ":`)
				version := strings.Trim(tag[i+1:], ` ":`)
				lastDep.Name = name
				lock[npmkey(name, version)] = lastDep
			}
			return
		}

		if strings.HasPrefix(line, "    ") {
			line = strings.TrimSpace(line)
			i := strings.Index(line, " ")
			if i == -1 {
				logs.Warnf("parse file %s line: %s fail", file.Relpath(), line)
				return
			}
			name := strings.Trim(line[:i], `"`)
			version := strings.Trim(line[i+1:], `"`)
			lastDep.Dependencies[name] = version
			return
		}

		if strings.HasPrefix(line, "  ") {
			line = strings.TrimSpace(line)
			if !strings.HasPrefix(line, "version") {
				return
			}
			lastDep.Version = strings.Trim(strings.TrimPrefix(line, "version"), `" `)
			return
		}

	})

	return lock
}

// ParsePackageJsonWithYarnLock 借助yarn.lock文件解析pacakge.json
func ParsePackageJsonWithYarnLock(pkgjson *PackageJson, yarnlock map[string]*YarnLock) *model.DepGraph {

	root := &model.DepGraph{Name: pkgjson.Name, Version: pkgjson.Version, Path: pkgjson.File.Relpath()}

	_dep := _depSet().LoadOrStore

	// 记录依赖
	for _, lock := range yarnlock {
		dep := _dep(lock.Name, lock.Version)
		for name, version := range lock.Dependencies {
			sub := yarnlock[npmkey(name, version)]
			if sub != nil {
				dep.AppendChild(_dep(sub.Name, sub.Version))
			}
		}
	}

	for name, version := range pkgjson.Dependencies {
		lock := yarnlock[npmkey(name, version)]
		if lock != nil {
			root.AppendChild(_dep(lock.Name, lock.Version))
		} else {
			root.AppendChild(&model.DepGraph{Name: name, Version: version})
		}
	}

	for name, version := range pkgjson.DevDependencies {
		lock := yarnlock[npmkey(name, version)]
		if lock != nil {
			dep := _dep(lock.Name, lock.Version)
			devdep := _dep(lock.Name, lock.Version, "dev")
			for _, c := range dep.Children {
				devdep.AppendChild(c)
			}
			root.AppendChild(devdep)
		} else {
			root.AppendChild(&model.DepGraph{Name: name, Version: version, Develop: true})
		}
	}

	return root
}
