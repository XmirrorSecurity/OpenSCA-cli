/*
 * @Description: 解析package.json文件
 * @Date: 2022-01-07 17:00:41
 */

package javascript

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"opensca/internal/bar"
	"opensca/internal/cache"
	"opensca/internal/enum/language"
	"opensca/internal/logs"
	"opensca/internal/srt"
	"path"
	"sort"

	"github.com/Masterminds/semver"
)

// package.json 文件结构
type PkgJson struct {
	Name    string            `json:"name"`
	Version string            `json:"version"`
	License string            `json:"license"`
	DevDeps map[string]string `json:"devDependencies"`
	Deps    map[string]string `json:"dependencies"`
}

// npm下载文件结构
type NpmJson struct {
	Time     map[string]string  `json:"time"`
	Versions map[string]PkgJson `json:"versions"`
}

/**
 * @description: 解析package.json
 * @param {*srt.DepTree} depRoot 依赖树节点
 * @param {*srt.FileData} file 文件数据
 * @return {[]*srt.DepTree} 解析出的依赖列表
 */
func parsePackage(depRoot *srt.DepTree, file *srt.FileData) (deps []*srt.DepTree) {
	deps = []*srt.DepTree{}
	pkg := PkgJson{}
	if err := json.Unmarshal(file.Data, &pkg); err != nil {
		logs.Error(err)
		return
	}
	pkgDep := depRoot
	pkgDep.Name = pkg.Name
	pkgDep.Version = srt.NewVersion(pkg.Version)
	pkgDep.AddLicense(pkg.License)
	pkgDep.Path = path.Join(depRoot.Path, path.Base(file.Name))
	deps = append(deps, pkgDep)
	// 依赖列表map[name]version
	depMap := map[string]string{}
	for name, version := range pkg.DevDeps {
		depMap[name] = version
	}
	for name, version := range pkg.Deps {
		depMap[name] = version
	}
	// 组件名排序后添加到deps
	names := []string{}
	for name := range depMap {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		version := depMap[name]
		dep := srt.NewDepTree(pkgDep)
		dep.Name = name
		dep.Version = srt.NewVersion(version)
		deps = append(deps, dep)
	}
	// 记录出现过的组件
	exist := map[string]struct{}{}
	// 搜索子依赖
	q := srt.NewQueue()
	exist[pkgDep.Name] = struct{}{}
	for _, child := range pkgDep.Children {
		exist[child.Name] = struct{}{}
		q.Push(child)
	}
	for !q.Empty() {
		node := q.Pop().(*srt.DepTree)
		for _, sub := range npmSimulation(node) {
			if _, ok := exist[sub.Name]; !ok {
				bar.Npm.Add(1)
				exist[sub.Name] = struct{}{}
				q.Push(sub)
			}
		}
	}
	return
}

/**
 * @description: 模拟npm获取详细依赖信息
 * @param {*srt.DepTree} dep 直接依赖
 * @return {[]*srt.DepTree} 子依赖列表，子依赖的路径及语言字段均已赋值
 */
func npmSimulation(dep *srt.DepTree) (subDeps []*srt.DepTree) {
	subDeps = []*srt.DepTree{}
	dep.Language = language.JavaScript
	// 获取依赖数据
	data := cache.LoadCache(dep.Dependency)
	if len(data) == 0 {
		url := fmt.Sprintf(`https://r.cnpmjs.org/%s`, dep.Name)
		if rep, err := http.Get(url); err != nil {
			logs.Error(err)
			return
		} else {
			if data, err = ioutil.ReadAll(rep.Body); err != nil {
				logs.Error(err)
			} else {
				cache.SaveCache(dep.Dependency, data)
			}
			rep.Body.Close()
		}
	}
	// 解析依赖信息
	npm := NpmJson{}
	json.Unmarshal(data, &npm)
	// 查找符合范围内的最大版本
	latestVersion := ""
	// 当前最大版本
	var max *semver.Version
	if constraint, err := semver.NewConstraint(dep.Version.Org); err == nil {
		for ver := range npm.Time {
			if v, err := semver.NewVersion(ver); err == nil {
				if constraint.Check(v) {
					if max == nil || max.LessThan(v) {
						max = v
						latestVersion = ver
					}
				}
			}
		}
	} else {
		logs.Warn(err)
	}
	if latestVersion == "" {
		return
	}
	info := npm.Versions[latestVersion]
	dep.Version = srt.NewVersion(latestVersion)
	if dep.Parent != nil {
		dep.Path = path.Join(dep.Parent.Path, dep.Dependency.String())
	}
	dep.AddLicense(info.License)
	// 解析子依赖
	names := []string{}
	for name := range info.Deps {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		sub := srt.NewDepTree(dep)
		sub.Name = name
		sub.Version = srt.NewVersion(info.Deps[name])
		sub.Path = path.Join(dep.Path, sub.Dependency.String())
		sub.Language = language.JavaScript
		subDeps = append(subDeps, sub)
	}
	return
}
