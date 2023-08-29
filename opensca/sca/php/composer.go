package php

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"

	"github.com/Masterminds/semver/v3"
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
	Packages []*ComposerPackage `json:"packages"`
}
type ComposerPackage struct {
	Name       string            `json:"name"`
	Version    string            `json:"version"`
	License    []string          `json:"license"`
	Require    map[string]string `json:"require"`
	RequireDev map[string]string `json:"require-dev"`
}

type ComposerRepo struct {
	Packages struct {
		Versions map[string]*ComposerPackage `json:"versions"`
	} `json:"packages"`
}

func readJson[T any](reader io.Reader) *T {
	var data T
	err := json.NewDecoder(reader).Decode(&data)
	if err != nil {
		return nil
	}
	return &data
}

func ParseComposerJsonWithLock(json *ComposerJson, lock *ComposerLock) *model.DepGraph {

	root := &model.DepGraph{Name: json.Name, Path: json.File.Path()}
	root.AppendLicense(json.License)

	_dep := model.NewDepGraphMap(nil, func(s ...string) *model.DepGraph { return &model.DepGraph{Name: s[0]} }).LoadOrStore

	// 第一次遍历记录依赖信息
	for _, pkg := range lock.Packages {
		dep := _dep(pkg.Name)
		for _, lic := range pkg.License {
			dep.AppendLicense(lic)
		}
	}

	// 第二次遍历记录依赖关系
	for _, pkg := range lock.Packages {
		dep := _dep(pkg.Name)
		for name := range pkg.Require {
			dep.AppendChild(_dep(name))
		}
		for name := range pkg.RequireDev {
			dep := _dep(name)
			if dep != nil {
				dep.Develop = true
				dep.AppendChild(dep)
			}
		}
	}

	// 记录直接依赖
	for name := range json.Require {
		root.AppendChild(_dep(name))
	}

	return root
}

func ParseComposerJsonWithOrigin(json *ComposerJson) *model.DepGraph {

	root := &model.DepGraph{Name: json.Name, Path: json.File.Path()}
	root.AppendLicense(json.License)

	_dep := model.NewDepGraphMap(nil, func(s ...string) *model.DepGraph { return &model.DepGraph{Name: s[0], Version: s[1]} }).LoadOrStore

	root.Expand = &ComposerPackage{
		Name:       json.License,
		Require:    json.Require,
		RequireDev: json.RequireDev,
	}

	root.ForEachNode(func(p, n *model.DepGraph) bool {

		pkg := n.Expand.(*ComposerPackage)
		for _, lic := range pkg.License {
			n.AppendLicense(lic)
		}

		for name, version := range pkg.Require {

			subpkg := phpOrigin(name, version)

			if subpkg == nil {
				n.AppendChild(_dep(name, version))
				continue
			}

			dep := _dep(subpkg.Name, subpkg.Version)
			dep.Expand = subpkg
			n.AppendChild(dep)
		}

		for name, version := range pkg.RequireDev {

			subpkg := phpOrigin(name, version)

			if subpkg == nil {
				dep := _dep(name, version)
				dep.Develop = true
				n.AppendChild(dep)
				continue
			}

			dep := _dep(subpkg.Name, subpkg.Version)
			dep.Expand = subpkg
			dep.Develop = true
			n.AppendChild(dep)
		}

		return true
	})

	root.ForEachNode(func(p, n *model.DepGraph) bool { n.Expand = nil; return true })

	return root
}

var phpOrigin = func(name, version string) *ComposerPackage {

	var origin *ComposerPackage

	findComposerJsonFromRepo := func(repo ComposerRepo) *ComposerPackage {
		vers := []string{}
		for v := range repo.Packages.Versions {
			vers = append(vers, v)
		}
		maxv := findMaxVersion(version, vers)
		return repo.Packages.Versions[maxv]
	}

	// 读取缓存
	path := cache.Path("", name, version, model.Lan_Php)
	cache.Load(path, func(reader io.Reader) {
		var repo ComposerRepo
		if err := json.NewDecoder(reader).Decode(&repo); err != nil {
			logs.Warn(err)
		} else {
			origin = findComposerJsonFromRepo(repo)
		}
	})

	if origin != nil {
		return origin
	}

	// 从composer仓库下载
	url := fmt.Sprintf("https://repo.packagist.org/p2/%s.json", name)
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
				return origin
			}

			reader := bytes.NewReader(data)
			var repo ComposerRepo
			if err := json.NewDecoder(reader).Decode(&repo); err != nil {
				logs.Warnf("unmarshal json from %s err: %s", url, err)
				return origin
			}

			reader.Seek(0, io.SeekStart)
			cache.Save(path, reader)
			origin = findComposerJsonFromRepo(repo)
		}
	}

	return origin
}

func RegisterNpmOrigin(origin func(name, version string) *ComposerPackage) {
	if origin != nil {
		phpOrigin = origin
	}
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
