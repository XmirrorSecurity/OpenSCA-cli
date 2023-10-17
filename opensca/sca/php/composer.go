package php

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/xmirrorsecurity/opensca-cli/opensca/common"
	"github.com/xmirrorsecurity/opensca-cli/opensca/logs"
	"github.com/xmirrorsecurity/opensca-cli/opensca/model"
	"github.com/xmirrorsecurity/opensca-cli/opensca/sca/cache"
)

type ComposerJson struct {
	Name       string            `json:"name"`
	License    string            `json:"license"`
	Require    map[string]string `json:"require"`
	RequireDev map[string]string `json:"require-dev"`
	File       *model.File       `json:"-"`
}

type ComposerLock struct {
	Packages    []*ComposerPackage `json:"packages"`
	PackagesDev []*ComposerPackage `json:"packages-dev"`
}
type ComposerPackage struct {
	Name       string            `json:"name"`
	Version    string            `json:"version"`
	License    []string          `json:"license"`
	Require    map[string]string `json:"require"`
	requireDev map[string]string
}

type ComposerRepo struct {
	Packages map[string][]*ComposerPackage `json:"packages"`
}

// skip 跳过非第三方组件
func skip(s string) bool {
	return !strings.Contains(s, "/")
}

func ParseComposerJsonWithLock(json *ComposerJson, lock *ComposerLock) *model.DepGraph {

	root := &model.DepGraph{Name: json.Name, Path: json.File.Relpath()}
	root.AppendLicense(json.License)

	_dep := model.NewDepGraphMap(nil, func(s ...string) *model.DepGraph { return &model.DepGraph{Name: s[0]} }).LoadOrStore

	// 第一次遍历记录依赖信息
	for _, pkg := range lock.Packages {
		dep := _dep(pkg.Name)
		dep.Version = pkg.Version
		for _, lic := range pkg.License {
			dep.AppendLicense(lic)
		}
	}
	for _, pkg := range lock.PackagesDev {
		dep := _dep(pkg.Name)
		dep.Version = pkg.Version
		dep.Develop = true
		for _, lic := range pkg.License {
			dep.AppendLicense(lic)
		}
	}

	// 第二次遍历记录依赖关系
	for _, pkg := range append(lock.Packages, lock.PackagesDev...) {
		dep := _dep(pkg.Name)
		for name := range pkg.Require {
			if skip(name) {
				continue
			}
			dep.AppendChild(_dep(name))
		}
	}

	// 记录直接依赖
	names := []string{}
	for name := range json.Require {
		names = append(names, name)
	}
	for name := range json.RequireDev {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if skip(name) {
			continue
		}
		root.AppendChild(_dep(name))
	}

	return root
}

func ParseComposerJsonWithOrigin(json *ComposerJson) *model.DepGraph {

	root := &model.DepGraph{Name: json.Name, Path: json.File.Relpath()}
	root.AppendLicense(json.License)

	_dep := model.NewDepGraphMap(nil, func(s ...string) *model.DepGraph { return &model.DepGraph{Name: s[0], Version: s[1]} }).LoadOrStore

	parseRequire := func(n *model.DepGraph, req map[string]string, dev bool) {

		names := []string{}
		for name := range req {
			names = append(names, name)
		}
		sort.Strings(names)

		for _, name := range names {

			if skip(name) {
				continue
			}

			version := req[name]
			subpkg := composerOrigin(name, version)

			if subpkg == nil {
				dep := _dep(name, version)
				dep.Develop = dev
				n.AppendChild(dep)
				continue
			}

			dep := _dep(subpkg.Name, subpkg.Version)
			if dep.Expand == nil {
				dep.Expand = subpkg
				dep.Develop = dev
			} else if !dev {
				dep.Develop = dev
			}
			n.AppendChild(dep)
		}
	}

	root.Expand = &ComposerPackage{
		Name:       json.Name,
		Require:    json.Require,
		requireDev: json.RequireDev,
	}

	root.ForEachNode(func(p, n *model.DepGraph) bool {

		if n.Expand == nil {
			return true
		}

		pkg := n.Expand.(*ComposerPackage)
		for _, lic := range pkg.License {
			n.AppendLicense(lic)
		}

		parseRequire(n, pkg.Require, false)
		parseRequire(n, pkg.requireDev, true)

		return true
	})

	root.ForEachNode(func(p, n *model.DepGraph) bool { n.Expand = nil; return true })

	return root
}

var composerOrigin = func(name, version string) *ComposerPackage {

	// 读取缓存
	var origin *ComposerPackage
	path := cache.Path("", name, version, model.Lan_Php)
	cache.Load(path, func(reader io.Reader) {
		origin = ReadComposerRepoJson(reader, name, version)
	})

	if origin != nil {
		return origin
	}

	// 从composer仓库下载
	common.DownloadUrlFromRepos(fmt.Sprintf("%s.json", name), func(repo common.RepoConfig, r io.Reader) {
		data, err := io.ReadAll(r)
		if err != nil {
			logs.Warn(err)
			return
		}
		reader := bytes.NewReader(data)
		origin = ReadComposerRepoJson(reader, name, version)
		reader.Seek(0, io.SeekStart)
		cache.Save(path, reader)
	}, common.RepoConfig{Url: "http://repo.packagist.org/p2"})

	return origin
}

func RegisterComposerOrigin(origin func(name, version string) *ComposerPackage) {
	if origin != nil {
		composerOrigin = origin
	}
}

func ReadComposerRepoJson(reader io.Reader, name, version string) *ComposerPackage {
	var repo ComposerRepo
	if err := json.NewDecoder(reader).Decode(&repo); err != nil {
		logs.Warnf("unmarshal %s err: %s", name, err)
	}
	vers := []string{}
	for _, packages := range repo.Packages {
		for _, pkg := range packages {
			vers = append(vers, pkg.Version)
		}
	}
	maxv := FindMaxVersion(version, vers)
	for _, packages := range repo.Packages {
		for _, pkg := range packages {
			if strings.EqualFold(pkg.Version, maxv) {
				pkg.Name = name
				return pkg
			}
		}
	}
	return nil
}

func FindMaxVersion(version string, versions []string) string {
	fix := func(s string) string {
		if i := strings.Index(s, "@"); i != -1 {
			s = s[:i]
		}
		return strings.TrimLeft(strings.TrimSpace(s), "vV")
	}
	var cs []*semver.Constraints
	for _, v := range strings.Split(version, "|") {
		c, err := semver.NewConstraint(fix(v))
		if err == nil {
			cs = append(cs, c)
		}
	}
	var vers []string
	for _, v := range versions {
		ver, err := semver.NewVersion(fix(v))
		if err != nil {
			continue
		}
		for _, c := range cs {
			if c.Check(ver) {
				vers = append(vers, v)
				break
			}
		}
	}
	sort.Slice(vers, func(i, j int) bool {
		v1, _ := semver.NewVersion(fix(vers[i]))
		v2, _ := semver.NewVersion(fix(vers[j]))
		return v1.Compare(v2) > 0
	})
	if len(vers) > 0 {
		return vers[0]
	}
	return version
}
