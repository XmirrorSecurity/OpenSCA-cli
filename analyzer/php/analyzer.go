/*
 * @Description: php解析器
 * @Date: 2021-11-26 14:39:49
 */

package php

import (
	"path"
	"util/enum/language"
	"util/filter"
	"util/model"
)

type Analyzer struct{}

// New 创建php解析器
func New() Analyzer {
	return Analyzer{}
}

// GetLanguage Get language of Analyzer
func (Analyzer) GetLanguage() language.Type {
	return language.Php
}

// CheckFile Check if it is a parsable file
func (Analyzer) CheckFile(filename string) bool {
	return filter.PhpComposerLock(filename) || filter.PhpComposer(filename)
}

// ParseFiles Parse the file
func (Analyzer) ParseFiles(files []*model.FileInfo) (deps []*model.DepTree) {
	deps = []*model.DepTree{}
	cpsMap := map[string]*model.FileInfo{}
	lockMap := map[string]*model.FileInfo{}
	for _, f := range files {
		if filter.PhpComposer(f.Name) {
			cpsMap[path.Dir(f.Name)] = f
		} else if filter.PhpComposerLock(f.Name) {
			lockMap[path.Dir(f.Name)] = f
		}
	}
	for _, f := range files {
		dep := model.NewDepTree(nil)
		dep.Path = f.Name
		if filter.PhpComposer(f.Name) {
			if _, ok := lockMap[path.Dir(f.Name)]; !ok {
				parseComposer(dep, f, true)
			}
		} else if filter.PhpComposerLock(f.Name) {
			if cps, ok := cpsMap[path.Dir(f.Name)]; !ok {
				parseComposerLock(dep, f, nil)
			} else {
				parseComposerLock(dep, f, parseComposer(dep, cps, false))
			}
		}
		deps = append(deps, dep)
	}
	return deps
}
