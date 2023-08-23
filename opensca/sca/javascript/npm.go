package javascript

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/xmirrorsecurity/opensca-cli/opensca/model"
	"github.com/xmirrorsecurity/opensca-cli/opensca/sca/cache"
	"github.com/xmirrorsecurity/opensca-cli/opensca/sca/filter"
)

type PackageJson struct {
	Name       string            `json:"name"`
	Version    string            `json:"version"`
	License    string            `json:"license"`
	DevDeps    map[string]string `json:"devDependencies"`
	Deps       map[string]string `json:"dependencies"`
	HomePage   string            `json:"homepage"`
	Repository map[string]string `json:"repository,omitempty"`
	File       *model.File       `json:"-"`
}

type NpmJson struct {
	Time     map[string]string       `json:"time"`
	Versions map[string]*PackageJson `json:"versions"`
}

type PackageLock struct {
	Name         string                     `json:"name"`
	Version      string                     `json:"version"`
	Dependencies map[string]*PackageLockDep `json:"dependencies"`
	File         *model.File                `json:"-"`
}
type PackageLockDep struct {
	name         string
	Version      string                     `json:"version"`
	Requires     map[string]string          `json:"requires"`
	Dependencies map[string]*PackageLockDep `json:"dependencies"`
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

	jsonMap := map[string]*PackageJson{}
	lockMap := map[string]*PackageLock{}
	nodeMap := map[string]map[string]*PackageJson{}

	// 将npm相关文件按上述方案分类
	for _, f := range files {
		if filter.JavaScriptPackageJson(f.Relpath) {
			var js *PackageJson
			f.OpenReader(func(reader io.Reader) {
				js = readJson[PackageJson](reader)
				if js == nil {
					return
				}
				if strings.Contains(f.Relpath, "node_modules") {
					if nodeMap[js.Name] == nil {
						nodeMap[js.Name] = map[string]*PackageJson{}
					}
					nodeMap[js.Name][js.Version] = js
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

func ParsePackageJsonWithLock(js *PackageJson, lock *PackageLock) *model.DepGraph {

	root := &model.DepGraph{Name: js.Name, Version: js.Version}

	if js.File != nil {
		root.Path = js.File.Relpath
	}

	// map[key]
	depMap := map[string]*model.DepGraph{}
	depNameMap := map[string]*model.DepGraph{}
	_dep := func(name, version string) *model.DepGraph {
		key := fmt.Sprintf("%s:%s", name, version)
		dep, ok := depMap[key]
		if !ok {
			dep = &model.DepGraph{Name: name, Version: version}
			depMap[key] = dep
		}
		return dep
	}

	// 记录依赖
	for name, lockDep := range lock.Dependencies {
		depNameMap[name] = _dep(name, lockDep.Version)
	}

	// 构建依赖关系
	for name := range js.Deps {

		lockDep := lock.Dependencies[name]
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

	for name := range js.Deps {
		root.AppendChild(depMap[name])
	}

	for name := range js.DevDeps {
		dep := depMap[name]
		dep.Develop = true
		root.AppendChild(dep)
	}

	return root
}

func ParsePackageJsonWithNode(js *PackageJson, nodeMap map[string]map[string]*PackageJson) *model.DepGraph {

	root := &model.DepGraph{Name: js.Name, Version: js.Version}

	if js.File != nil {
		root.Path = js.File.Relpath
	}

	depMap := map[string]*model.DepGraph{}
	_dep := func(name, version string) *model.DepGraph {
		key := fmt.Sprintf("%s:%s", name, version)
		dep, ok := depMap[key]
		if !ok {
			dep = &model.DepGraph{Name: name, Version: version}
			depMap[key] = dep
		}
		return dep
	}

	findDep := func(name, version string) *model.DepGraph {
		if len(nodeMap) > 0 {
			// 从node_modules中查找
			versions := []string{}
			for version := range nodeMap[name] {
				versions = append(versions, version)
			}
			maxv := findMaxVersion(version, versions)
			dep := _dep(name, maxv)
			if dep.Expand == nil {
				dep.Expand = nodeMap[name][maxv]
			}
			return dep
		} else {
			// 从外部数据源下载
			ojs := npmOrigin(name, version)
			dep := _dep(ojs.Name, ojs.Version)
			if dep.Expand == nil {
				dep.Expand = ojs
			}
			return dep
		}
	}

	root.Expand = js
	root.ForEachPath(func(p, n *model.DepGraph) bool {
		njs := n.Expand.(*PackageJson)
		for name, version := range njs.Deps {
			n.AppendChild(findDep(name, version))
		}
		for name, version := range njs.DevDeps {
			dep := findDep(name, version)
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
