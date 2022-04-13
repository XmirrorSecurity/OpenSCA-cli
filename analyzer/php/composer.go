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
	"sort"
	"strings"
	"util/bar"
	"util/cache"
	"util/enum/language"
	"util/logs"
	"util/model"

	"github.com/Masterminds/semver/v3"
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

// parseComposer parse composer.json
func parseComposer(depRoot *model.DepTree, file *model.FileData) (deps []*model.DepTree) {
	deps = []*model.DepTree{}
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
		dep := model.NewDepTree(depRoot)
		dep.Name = name
		dep.Version = model.NewVersion(requires[name])
		if composer.Name == "" {
			deps = append(deps, dep)
		}
	}
	// composer simulation
	exist := map[string]struct{}{}
	exist[depRoot.Name] = struct{}{}
	q := model.NewQueue()
	for _, child := range depRoot.Children {
		exist[child.Name] = struct{}{}
		q.Push(child)
	}
	for !q.Empty() {
		node := q.Pop().(*model.DepTree)
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

// composerSimulation composer simulation
func composerSimulation(dep *model.DepTree) (subDeps []*model.DepTree) {
	subDeps = []*model.DepTree{}
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
						dep.Version = model.NewVersion(info.Version)
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
							sub := model.NewDepTree(dep)
							sub.Name = name
							sub.Version = model.NewVersion(requires[name])
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
