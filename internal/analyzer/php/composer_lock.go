/*
 * @Descripation: parse composer.lock file
 * @Date: 2021-11-26 14:50:06
 */

package php

import (
	"encoding/json"
	"opensca/internal/logs"
	"opensca/internal/srt"
)

// composer.lock
type ComposerLock struct {
	Pkgs []struct {
		Name    string            `json:"name"`
		Version string            `json:"version"`
		Require map[string]string `json:"require"`
	} `json:"packages"`
}

/**
 * @description: parse composer.lock
 * @param {*srt.DepTree} depRoot dependency
 * @param {*srt.FileData} file composer.lock file data
 * @return {[]*srt.DepTree} dependencies list
 */
func parseComposerLock(depRoot *srt.DepTree, file *srt.FileData) (deps []*srt.DepTree) {
	deps = []*srt.DepTree{}
	lock := ComposerLock{}
	if err := json.Unmarshal(file.Data, &lock); err != nil {
		logs.Error(err)
		return
	}
	// dependencies info
	// map[name]DepTree
	depMap := map[string]*srt.DepTree{}
	for _, cps := range lock.Pkgs {
		dep := srt.NewDepTree(nil)
		dep.Name = cps.Name
		dep.Version = srt.NewVersion(cps.Version)
		depMap[cps.Name] = dep
	}
	// build dependency tree
	for _, cps := range lock.Pkgs {
		for n := range cps.Require {
			if sub, ok := depMap[n]; ok && sub.Parent == nil {
				dep := depMap[cps.Name]
				sub.Parent = dep
				dep.Children = append(dep.Children, sub)
			}
		}
	}
	// move direct dependices under the root
	for _, cps := range lock.Pkgs {
		dep := depMap[cps.Name]
		if dep.Parent == nil {
			dep.Parent = depRoot
			depRoot.Children = append(depRoot.Children, dep)
		}
		deps = append(deps, dep)
	}
	return
}
