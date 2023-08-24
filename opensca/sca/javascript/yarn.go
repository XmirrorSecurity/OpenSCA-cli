package javascript

import (
	"strings"

	"github.com/xmirrorsecurity/opensca-cli/opensca/logs"
	"github.com/xmirrorsecurity/opensca-cli/opensca/model"
)

type YarnLock struct {
	Name         string
	Version      string
	Dependencies map[string]string
}

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
				i := strings.Index(tag, "@")
				if i != -1 {
					logs.Warnf("parse file %s line: %s fail", file.Relpath, line)
					continue
				}
				name := strings.Trim(tag[:i], `":`)
				version := strings.Trim(tag[i+1:], `":`)
				lastDep.Name = name
				lock[npmkey(name, version)] = lastDep
			}
			return
		}

		if strings.HasPrefix(line, "    ") {
			line = strings.TrimSpace(line)
			i := strings.Index(line, " ")
			if i != -1 {
				logs.Warnf("parse file %s line: %s fail", file.Relpath, line)
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

// parseYarn 解析yarn文件
func ParsePackageJsonWithYarnLock(js *PackageJson, yarn map[string]*YarnLock) *model.DepGraph {

	root := &model.DepGraph{Name: js.Name, Version: js.Version}
	root.AppendFile(js.File)

	_dep := (&depSet{}).Dep

	// 记录依赖
	for _, lock := range yarn {
		dep := _dep(lock.Name, lock.Version)
		for name, version := range lock.Dependencies {
			sub := yarn[npmkey(name, version)]
			dep.AppendChild(_dep(sub.Name, sub.Version))
		}
	}

	for name, version := range js.Dependencies {
		lock := yarn[npmkey(name, version)]
		if lock != nil {
			root.AppendChild(_dep(lock.Name, lock.Version))
		} else {
			root.AppendChild(&model.DepGraph{Name: name, Version: version})
		}
	}

	for name, version := range js.DevDependencies {
		lock := yarn[npmkey(name, version)]
		if lock != nil {
			dep := _dep(lock.Name, lock.Version)
			dep.Develop = true
			root.AppendChild(dep)
		} else {
			root.AppendChild(&model.DepGraph{Name: name, Version: version, Develop: true})
		}
	}

	return root
}
