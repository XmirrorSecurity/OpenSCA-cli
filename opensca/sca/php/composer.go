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
	Packages []*ComposerPackage `json:"packages"`
}
type ComposerPackage struct {
	Name    string            `json:"name"`
	Version string            `json:"version"`
	License []string          `json:"license"`
	Require map[string]string `json:"require"`
}

type ComposerRepo struct {
	Packages map[string][]*ComposerPackage `json:"packages"`
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
		Name:    json.License,
		Require: json.Require,
	}

	root.ForEachNode(func(p, n *model.DepGraph) bool {

		if n.Expand == nil {
			return true
		}

		pkg := n.Expand.(*ComposerPackage)
		for _, lic := range pkg.License {
			n.AppendLicense(lic)
		}

		for name, version := range pkg.Require {

			if name == "php" {
				continue
			}

			subpkg := composerOrigin(name, version)

			if subpkg == nil {
				n.AppendChild(_dep(name, version))
				continue
			}

			dep := _dep(subpkg.Name, subpkg.Version)
			dep.Expand = subpkg
			n.AppendChild(dep)
		}

		return true
	})

	// for name, version := range json.RequireDev {
	// 	if name == "php" {
	// 		continue
	// 	}
	// 	subpkg := composerOrigin(name, version)
	// 	if subpkg == nil {
	// 		dep := _dep(name, version)
	// 		dep.Develop = true
	// 		root.AppendChild(dep)
	// 		continue
	// 	}
	// 	dep := _dep(subpkg.Name, subpkg.Version)
	// 	dep.Develop = true
	// 	root.AppendChild(dep)
	// }

	root.ForEachNode(func(p, n *model.DepGraph) bool { n.Expand = nil; return true })

	return root
}

var composerOrigin = func(name, version string) *ComposerPackage {

	findComposerJsonFromRepo := func(repo ComposerRepo) *ComposerPackage {
		vers := []string{}
		for _, packages := range repo.Packages {
			for _, pkg := range packages {
				vers = append(vers, pkg.Version)
			}
		}
		maxv := findMaxVersion(version, vers)
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

	// 读取缓存
	var origin *ComposerPackage
	path := cache.Path("", name, version, model.Lan_Php)
	if cache.Load(path, func(reader io.Reader) {
		var repo ComposerRepo
		if err := json.NewDecoder(reader).Decode(&repo); err != nil {
			logs.Warnf("unmarshal %s err: %s", name, err)
		}
		origin = findComposerJsonFromRepo(repo)
	}) {
		return origin
	}

	// 从composer仓库下载
	for _, repo := range defaultComposerRepo {

		url := fmt.Sprintf("%s/%s.json", strings.TrimRight(repo.Url, "/"), name)
		resp, err := common.HttpClient.Get(url)
		if err != nil {
			logs.Warnf("download %s err: %s", url, err)
			continue
		}

		if resp.StatusCode != 200 {
			logs.Warnf("code:%d url:%s", resp.StatusCode, url)
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
			continue
		}

		logs.Infof("code:%d url:%s", resp.StatusCode, url)

		data, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			logs.Warn(err)
			continue
		}

		reader := bytes.NewReader(data)
		var repo ComposerRepo
		if err := json.NewDecoder(reader).Decode(&repo); err != nil {
			logs.Warnf("unmarshal json from %s err: %s", url, err)
			// continue
		}

		reader.Seek(0, io.SeekStart)
		cache.Save(path, reader)

		origin = findComposerJsonFromRepo(repo)
		if origin != nil {
			return origin
		}
	}

	return nil
}

func RegisterComposerOrigin(origin func(name, version string) *ComposerPackage) {
	if origin != nil {
		composerOrigin = origin
	}
}

func findMaxVersion(version string, versions []string) string {
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
		return v1.Compare(v2) < 0
	})
	if len(vers) > 0 {
		return vers[0]
	}
	return version
}
