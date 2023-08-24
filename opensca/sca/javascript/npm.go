package javascript

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/xmirrorsecurity/opensca-cli/opensca/logs"
	"github.com/xmirrorsecurity/opensca-cli/opensca/model"
	"github.com/xmirrorsecurity/opensca-cli/opensca/sca/cache"
	"github.com/xmirrorsecurity/opensca-cli/opensca/sca/filter"
)

type PackageJson struct {
	Name            string            `json:"name"`
	Version         string            `json:"version"`
	License         string            `json:"license"`
	DevDependencies map[string]string `json:"devDependencies"`
	Dependencies    map[string]string `json:"dependencies"`
	File            *model.File       `json:"-"`
}

type NpmJson struct {
	Time     map[string]string       `json:"time"`
	Versions map[string]*PackageJson `json:"versions"`
}

type PackageLock struct {
	Name         string                     `json:"name"`
	LockVersion  int                        `json:"lockVersion"`
	Version      string                     `json:"version"`
	Dependencies map[string]*PackageLockDep `json:"dependencies"`
	Packages     map[string]*PackageJson    `json:"packages"`
}
type PackageLockDep struct {
	name         string
	Version      string                     `json:"version"`
	Requires     map[string]string          `json:"requires"`
	Dependencies map[string]*PackageLockDep `json:"dependencies"`
}

func npmkey(name, version string) string {
	return fmt.Sprintf("%s:%s", name, version)
}

type depSet struct {
	m map[string]*model.DepGraph
}

func (s *depSet) Dep(name, version string) *model.DepGraph {
	if s.m == nil {
		s.m = map[string]*model.DepGraph{}
	}
	key := npmkey(name, version)
	dep, ok := s.m[key]
	if !ok {
		dep = &model.DepGraph{Name: name, Version: version}
		s.m[key] = dep
	}
	return dep
}

func readJson[T any](reader io.Reader) *T {
	var data T
	err := json.NewDecoder(reader).Decode(data)
	if err != nil {
		return nil
	}
	return &data
}

var npmOrigin = func(name, version string) *PackageJson {

	var origin *PackageJson

	// 读取缓存
	path := cache.Path("", name, version, model.Lan_JavaScript)
	cache.Load(path, func(reader io.Reader) {
		npm := readJson[NpmJson](reader)
		vers := []string{}
		for v := range npm.Versions {
			vers = append(vers, v)
		}
		origin = npm.Versions[findMaxVersion(version, vers)]
	})

	if origin != nil {
		return origin
	}

	// 从npm仓库下载
	url := fmt.Sprintf(`https://r.cnpmjs.org/%s`, name)
	if rep, err := http.Get(url); err == nil {
		defer rep.Body.Close()
		if rep.StatusCode != 200 {
			logs.Warnf("code:%d url:%s", rep.StatusCode, url)
			io.Copy(io.Discard, rep.Body)
		} else {
			logs.Infof("code:%d url:%s", rep.StatusCode, url)
			data, err := io.ReadAll(rep.Body)
			if err == nil {
				reader := bytes.NewReader(data)
				cache.Save(path, reader)
				reader.Seek(0, io.SeekStart)
				npm := readJson[NpmJson](reader)
				vers := []string{}
				for v := range npm.Versions {
					vers = append(vers, v)
				}
				origin = npm.Versions[findMaxVersion(version, vers)]
			}
		}
	}

	return origin
}

func RegisterNpmOrigin(origin func(name, version string) *PackageJson) {
	if origin != nil {
		npmOrigin = origin
	}
}

// files: package.json package-lock.json
func ParseNpm(files []*model.File) []*model.DepGraph {

	var root []*model.DepGraph

	// map[name]
	jsonMap := map[string]*PackageJson{}
	// map[name]
	lockMap := map[string]*PackageLock{}
	// map[path]
	nodeMap := map[string]*PackageJson{}

	// 将npm相关文件按上述方案分类
	for _, f := range files {
		if filter.JavaScriptPackageJson(f.Relpath) {
			var js *PackageJson
			f.OpenReader(func(reader io.Reader) {
				js = readJson[PackageJson](reader)
				if js == nil {
					return
				}
				js.File = f
				if strings.Contains(f.Relpath, "node_modules") {
					nodeMap[f.Relpath] = js
				} else {
					jsonMap[js.Name] = js
				}
			})
		}
		if filter.JavaScriptPackageLock(f.Relpath) {
			f.OpenReader(func(reader io.Reader) {
				lock := readJson[PackageLock](reader)
				if lock == nil {
					return
				}
				lockMap[lock.Name] = lock
			})
		}
	}

	// 遍历非node_modules下的package.json
	for dir, js := range jsonMap {
		var dep *model.DepGraph
		if lock, ok := lockMap[dir]; ok {
			dep = ParsePackageJsonWithLock(js, lock)
		} else {
			dep = ParsePackageJsonWithNode(js, nodeMap)
		}
		root = append(root, dep)
	}

	return root
}

func ParsePackageJsonWithLockV3(js *PackageJson, lock *PackageLock) *model.DepGraph {

	if lock.LockVersion != 3 {
		return nil
	}

	root := &model.DepGraph{Name: js.Name, Version: js.Version}
	if js.File != nil {
		root.Path = js.File.Relpath
	}

	type expand struct {
		js   *PackageJson
		path string
	}
	root.Expand = expand{js: js, path: ""}

	_dep := (&depSet{}).Dep

	findDep := func(name, basedir string) *model.DepGraph {
		jspath, subjs := findFromNodeModules(name, basedir, lock.Packages)
		if subjs == nil {
			return nil
		}
		dep := _dep(subjs.Name, subjs.Version)
		if dep.Expand == nil {
			dep.Expand = expand{
				path: jspath,
				js:   subjs,
			}
		}
		return dep
	}

	root.ForEachPath(func(p, n *model.DepGraph) bool {

		njs := n.Expand.(expand)
		if njs.js == nil {
			return false
		}

		for name := range njs.js.Dependencies {
			n.AppendChild(findDep(name, njs.path))
		}

		for name := range njs.js.DevDependencies {
			dep := findDep(name, njs.path)
			dep.Develop = true
			n.AppendChild(dep)
		}

		return true
	})
	return nil
}

func ParsePackageJsonWithLock(js *PackageJson, lock *PackageLock) *model.DepGraph {

	if lock.LockVersion == 3 {
		return ParsePackageJsonWithLockV3(js, lock)
	}

	root := &model.DepGraph{Name: js.Name, Version: js.Version}
	if js.File != nil {
		root.Path = js.File.Relpath
	}

	// map[key]
	depNameMap := map[string]*model.DepGraph{}
	_dep := (&depSet{}).Dep

	// 记录依赖
	for name, lockDep := range lock.Dependencies {
		depNameMap[name] = _dep(name, lockDep.Version)
	}

	// 构建依赖关系
	for name, lockDep := range lock.Dependencies {
		lockDep.name = name
		q := []*PackageLockDep{lockDep}
		for len(q) > 0 {
			n := q[0]
			q = q[1:]

			dep := _dep(n.name, n.Version)

			for name, sub := range n.Dependencies {
				sub.name = name
				q = append(q, sub)
				dep.AppendChild(_dep(name, sub.Version))
			}

			for name := range n.Requires {
				dep.AppendChild(depNameMap[name])
			}

		}
	}

	for name := range js.Dependencies {
		root.AppendChild(depNameMap[name])
	}

	for name := range js.DevDependencies {
		dep := depNameMap[name]
		dep.Develop = true
		root.AppendChild(dep)
	}

	return root
}

func ParsePackageJsonWithNode(js *PackageJson, nodeMap map[string]*PackageJson) *model.DepGraph {

	root := &model.DepGraph{Name: js.Name, Version: js.Version}

	if js.File != nil {
		root.Path = js.File.Relpath
	}

	_dep := (&depSet{}).Dep

	root.Expand = js

	findDep := func(name, version, basedir string) *model.DepGraph {
		var subjs *PackageJson
		if len(nodeMap) > 0 {
			// 从node_modules中查找
			_, subjs = findFromNodeModules(name, basedir, nodeMap)
		} else {
			// 从外部数据源下载
			subjs = npmOrigin(name, version)
		}
		if subjs == nil {
			return nil
		}
		dep := _dep(subjs.Name, subjs.Version)
		if dep.Expand == nil {
			dep.Expand = subjs
		}
		return dep
	}

	root.ForEachPath(func(p, n *model.DepGraph) bool {

		njs := n.Expand.(*PackageJson)
		var basedir string
		if njs.File != nil {
			basedir = njs.File.Relpath
		}

		for name, version := range njs.Dependencies {
			n.AppendChild(findDep(name, version, basedir))
		}

		for name, version := range njs.DevDependencies {
			dep := findDep(name, version, basedir)
			dep.Develop = true
			n.AppendChild(dep)
		}

		return true
	})
	root.ForEachNode(func(p, n *model.DepGraph) bool { n.Expand = nil; return true })

	return root
}

func findMaxVersion(version string, versions []string) string {
	c, err := semver.NewConstraint(version)
	if err != nil {
		return version
	}
	vers := []*semver.Version{}
	for _, v := range versions {
		ver, err := semver.NewVersion(v)
		if err != nil {
			continue
		}
		if c.Check(ver) {
			vers = append(vers, ver)
		}
	}
	sort.Slice(vers, func(i, j int) bool {
		return vers[i].Compare(vers[j]) > 0
	})
	if len(vers) > 0 {
		return vers[0].String()
	}
	return version
}

func findFromNodeModules(name, basedir string, nodePathMap map[string]*PackageJson) (jspath string, js *PackageJson) {
	dir := basedir
	for {
		jspath = path.Join(dir, "node_modules", name)
		if js, ok := nodePathMap[jspath]; ok {
			return jspath, js
		}
		i := strings.LastIndex(dir, "node_modules")
		if i == -1 {
			break
		}
		dir = dir[:i]
	}
	return
}
