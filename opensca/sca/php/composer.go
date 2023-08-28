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
	Packages []*ComposerLockPackage `json:"packages"`
}
type ComposerLockPackage struct {
	Name       string            `json:"name"`
	Version    string            `json:"version"`
	License    []string          `json:"license"`
	Require    map[string]string `json:"require"`
	RequireDev map[string]string `json:"require-dev"`
}

type ComposerRepo struct {
	Packages map[string][]*ComposerJson `json:"packages"`
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

	_dep := model.NewDepGraphMap(func(s ...string) string { return s[0] }, func(s ...string) *model.DepGraph { return &model.DepGraph{Name: s[0]} }).LoadOrStore

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

var phpOrigin = func(name, version string) *ComposerJson {

	var origin *ComposerJson

	// 读取缓存
	path := cache.Path("", name, version, model.Lan_Php)
	cache.Load(path, func(reader io.Reader) {
		repo := readJson[ComposerRepo](reader)
		origin = findComposerJsonFromRepo(repo, version)
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
			if err == nil {
				reader := bytes.NewReader(data)
				cache.Save(path, reader)
				reader.Seek(0, io.SeekStart)
				repo := readJson[ComposerRepo](reader)
				origin = findComposerJsonFromRepo(repo, version)
			}
		}
	}

	return origin
}

func RegisterNpmOrigin(origin func(name, version string) *ComposerJson) {
	if origin != nil {
		phpOrigin = origin
	}
}

func findComposerJsonFromRepo(repo *ComposerRepo, version string) *ComposerJson {
	vers := []string{}
	for v := range repo.Packages {
		vers = append(vers, v)
	}
	maxv := findMaxVersion(version, vers)
	origin := repo.Packages[maxv][0]
	return origin
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
