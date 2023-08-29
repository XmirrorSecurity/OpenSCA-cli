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
	"github.com/xmirrorsecurity/opensca-cli/opensca/logs"
	"github.com/xmirrorsecurity/opensca-cli/opensca/model"
	"github.com/xmirrorsecurity/opensca-cli/opensca/sca/cache"
)

type PackageJson struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	// License              string            `json:"license"`
	Develop              bool              `json:"dev"` // lock v3
	Dependencies         map[string]string `json:"dependencies"`
	DevDependencies      map[string]string `json:"devDependencies"`
	PeerDependencies     map[string]string `json:"peerDependencies"`
	PeerDependenciesMeta map[string]struct {
		Optional bool `json:"optional"`
	} `json:"peerDependenciesMeta"`
	File *model.File `json:"-"`
}

type NpmJson struct {
	Time     map[string]string       `json:"time"`
	Versions map[string]*PackageJson `json:"versions"`
}

type PackageLock struct {
	Name            string                     `json:"name"`
	LockfileVersion int                        `json:"lockfileVersion"`
	Version         string                     `json:"version"`
	Dependencies    map[string]*PackageLockDep `json:"dependencies"`
	Packages        map[string]*PackageJson    `json:"packages"`
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

func _depSet() *model.DepGraphMap {
	return model.NewDepGraphMap(func(s ...string) string {
		return fmt.Sprintf("%s:%s", s[0], s[1])
	}, func(s ...string) *model.DepGraph {
		return &model.DepGraph{
			Name:    s[0],
			Version: s[1],
		}
	})
}

func readJson[T any](reader io.Reader) *T {
	var data T
	err := json.NewDecoder(reader).Decode(&data)
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
			if err != nil {
				logs.Warn(err)
			}

			reader := bytes.NewReader(data)
			var npm NpmJson
			if err := json.NewDecoder(reader).Decode(&npm); err != nil {
				logs.Warnf("unmarshal json from %s err: %s", url, err)
				return origin
			}

			vers := []string{}
			for v := range npm.Versions {
				vers = append(vers, v)
			}
			origin = npm.Versions[findMaxVersion(version, vers)]

			reader.Seek(0, io.SeekStart)
			cache.Save(path, reader)
		}
	}

	return origin
}

func RegisterNpmOrigin(origin func(name, version string) *PackageJson) {
	if origin != nil {
		npmOrigin = origin
	}
}

func ParsePackageJsonWithNode(pkgjson *PackageJson, nodeMap map[string]*PackageJson) *model.DepGraph {

	root := &model.DepGraph{Name: pkgjson.Name, Version: pkgjson.Version, Path: pkgjson.File.Path()}
	// root.AppendLicense(pkgjson.License)

	_dep := _depSet().LoadOrStore

	root.Expand = pkgjson

	findDep := func(name, version, basedir string) *model.DepGraph {
		var subjs *PackageJson
		if len(nodeMap) > 0 {
			// 从node_modules中查找
			_, subjs = findFromNodeModules(name, basedir, nodeMap)
		} else {
			// 从外部数据源下载
			subjs = npmOrigin(name, version)
			if subjs != nil {
				// 忽略开发环境依赖
				subjs.DevDependencies = nil
			}
		}
		if subjs == nil {
			return nil
		}
		dep := _dep(subjs.Name, subjs.Version)
		if dep.Expand == nil {
			// dep.AppendLicense(subjs.License)
			dep.Expand = subjs
		}
		return dep
	}

	root.ForEachPath(func(p, n *model.DepGraph) bool {

		njs := n.Expand.(*PackageJson)
		basedir := njs.File.Path()

		for name, version := range njs.Dependencies {
			n.AppendChild(findDep(name, version, basedir))
		}

		for name, version := range njs.DevDependencies {
			dep := findDep(name, version, basedir)
			if dep != nil {
				dep.Develop = true
				n.AppendChild(dep)
			}
		}

		return true
	})
	root.ForEachNode(func(p, n *model.DepGraph) bool { n.Expand = nil; return true })

	return root
}

func ParsePackageJsonWithLock(pkgjson *PackageJson, pkglock *PackageLock) *model.DepGraph {

	if pkglock.LockfileVersion == 3 {
		return ParsePackageJsonWithLockV3(pkgjson, pkglock)
	}

	root := &model.DepGraph{Name: pkgjson.Name, Version: pkgjson.Version, Path: pkgjson.File.Path()}
	// root.AppendLicense(pkgjson.License)

	// map[key]
	depNameMap := map[string]*model.DepGraph{}
	_dep := _depSet().LoadOrStore

	// 记录依赖
	for name, lockDep := range pkglock.Dependencies {
		dep := _dep(name, lockDep.Version)
		depNameMap[name] = dep
	}

	// 构建依赖关系
	for name, lockDep := range pkglock.Dependencies {
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

	for name := range pkgjson.Dependencies {
		root.AppendChild(depNameMap[name])
	}

	for name := range pkgjson.DevDependencies {
		dep := depNameMap[name]
		if dep != nil {
			dep.Develop = true
			root.AppendChild(dep)
		}
	}

	return root
}

func ParsePackageJsonWithLockV3(pkgjson *PackageJson, pkglock *PackageLock) *model.DepGraph {

	if pkglock.LockfileVersion != 3 {
		return nil
	}

	for jspath, js := range pkglock.Packages {
		if js.File == nil {
			js.File = &model.File{Relpath: jspath}
		}
	}

	root := &model.DepGraph{Name: pkgjson.Name, Version: pkgjson.Version, Path: pkgjson.File.Path()}
	// root.AppendLicense(pkgjson.License)

	type expand struct {
		js   *PackageJson
		path string
	}
	root.Expand = expand{js: pkgjson, path: ""}

	_dep := model.NewDepGraphMap(nil, func(s ...string) *model.DepGraph { return &model.DepGraph{Name: s[0], Version: s[1]} }).LoadOrStore

	findDep := func(name, basedir string) *model.DepGraph {
		jspath, subjs := findFromNodeModules(name, basedir, pkglock.Packages)
		if subjs == nil {
			return nil
		}
		dep := _dep(name, subjs.Version, subjs.File.Path())
		// dep.AppendLicense(subjs.License)
		if dep.Expand == nil {
			dep.Develop = subjs.Develop
			dep.Expand = expand{
				path: jspath,
				js:   subjs,
			}
		} else {
			if !subjs.Develop {
				dep.Develop = false
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

		for name := range njs.js.PeerDependencies {
			if meta, ok := njs.js.PeerDependenciesMeta[name]; ok {
				if meta.Optional {
					continue
				}
			}
			n.AppendChild(findDep(name, njs.path))
		}

		for name := range njs.js.DevDependencies {
			dep := findDep(name, njs.path)
			if dep != nil {
				dep.Develop = true
				n.AppendChild(dep)
			}
		}

		return true
	})

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
	const node_modules = "node_modules"
	paths := strings.Split(basedir, "/")
	for i := range paths {
		tail := len(paths) - i
		dirs := paths[:tail]
		if paths[tail-1] != node_modules {
			dirs = append(dirs, node_modules)
		}
		dirs = append(dirs, name)
		jspath = strings.TrimLeft(strings.Join(dirs, "/"), "/")
		if js, ok := nodePathMap[jspath]; ok {
			return jspath, js
		}
	}
	return
}
