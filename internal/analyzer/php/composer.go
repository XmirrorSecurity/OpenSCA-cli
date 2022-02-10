/*
 * @Description: parse composer_lock file
 * @Date: 2022-01-17 14:28:58
 */

package php

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
	"sort"
	"strings"

	"github.com/Masterminds/semver"
)

type Composer struct {
	Name       string            `json:"name"`
	License    string            `json:"license"`
	Require    map[string]string `json:"require"`
	RequireDev map[string]string `json:"require-dev"`
}

type ComposerRepo struct {
	Pkgs map[string][]struct {
		Version string            `json:"version"`
		Require map[string]string `json:"require"`
		// RequireDev map[string]string `json:"require-dev"`
	} `json:"packages"`
}

/**
 * @description: parse composer.json
 * @param {*srt.DepTree} depRoot dependency
 * @param {*srt.FileData} file composer.json file data
 * @return {[]*srt.DepTree} dependencies list
 */
func parseComposer(depRoot *srt.DepTree, file *srt.FileData) (deps []*srt.DepTree) {
	deps = []*srt.DepTree{}
	composer := Composer{}
	if err := json.Unmarshal(file.Data, &composer); err != nil {
		logs.Warn(err)
	}
	// set name
	if composer.Name != "" {
		depRoot.Name = composer.Name
		deps = append(deps, depRoot)
	}
	// add license
	if composer.License != "" {
		depRoot.AddLicense(composer.License)
	}
	// parse direct dependency
	requires := map[string]string{}
	for name, version := range composer.Require {
		requires[name] = version
	}
	for name, version := range composer.RequireDev {
		requires[name] = version
	}
	names := []string{}
	for name := range requires {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if strings.EqualFold(name, "php") {
			continue
		}
		dep := srt.NewDepTree(depRoot)
		dep.Name = name
		dep.Version = srt.NewVersion(requires[name])
		if composer.Name == "" {
			deps = append(deps, dep)
		}
	}
	// composer simulation
	exist := map[string]struct{}{}
	exist[depRoot.Name] = struct{}{}
	q := srt.NewQueue()
	for _, child := range depRoot.Children {
		exist[child.Name] = struct{}{}
		q.Push(child)
	}
	for !q.Empty() {
		node := q.Pop().(*srt.DepTree)
		for _, sub := range composerSimulation(node) {
			if _, ok := exist[sub.Name]; !ok {
				bar.Composer.Add(1)
				exist[sub.Name] = struct{}{}
				q.Push(sub)
			}
		}
	}
	return
}

/**
 * @description: composer simulation
 * @param {*srt.DepTree} dep dependency infomation
 * @return {[]*srt.DepTree} indirect dependencies
 */
func composerSimulation(dep *srt.DepTree) (subDeps []*srt.DepTree) {
	subDeps = []*srt.DepTree{}
	dep.Language = language.Php
	data := cache.LoadCache(dep.Dependency)
	if len(data) == 0 && dep.Name != "" {
		url := fmt.Sprintf("https://repo.packagist.org/p2/%s.json", dep.Name)
		rep, err := http.Get(url)
		if err != nil {
			logs.Warn(err)
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
	constraints := []*semver.Constraints{}
	for _, constraint := range strings.Split(dep.Version.Org, "|") {
		constraint = strings.TrimSpace(strings.ReplaceAll(constraint, "*", "x"))
		if c, err := semver.NewConstraint(constraint); err == nil {
			constraints = append(constraints, c)
		}
	}

	repo := ComposerRepo{}
	// ignore error
	_ = json.Unmarshal(data, &repo)

	// select version
	for _, infos := range repo.Pkgs {
		for _, info := range infos {
			if v, err := semver.NewVersion(info.Version); err == nil {
				for _, c := range constraints {
					if c.Check(v) {
						// add indirect dependencies
						dep.Version = srt.NewVersion(info.Version)
						requires := map[string]string{}
						for name, version := range info.Require {
							requires[name] = version
						}
						names := []string{}
						for name := range requires {
							names = append(names, name)
						}
						sort.Strings(names)
						for _, name := range names {
							if strings.EqualFold(name, "php") {
								continue
							}
							sub := srt.NewDepTree(dep)
							sub.Name = name
							sub.Version = srt.NewVersion(requires[name])
							subDeps = append(subDeps, sub)
						}
						return
					}
				}
			}
		}
	}
	return
}
