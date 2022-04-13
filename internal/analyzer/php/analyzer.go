/*
 * @Descripation: php解析器
 * @Date: 2021-11-26 14:39:49
 */

package php

import (
	"opensca/internal/enum/language"
	"opensca/internal/filter"
	"opensca/internal/srt"
	"sort"
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

// FilterFile filters the files that the current parser needs to parse
func (a Analyzer) FilterFile(dirRoot *srt.DirTree, depRoot *srt.DepTree) (files []*srt.FileData) {
	files = []*srt.FileData{}
	// 筛选需要解析的文件
	for _, f := range dirRoot.Files {
		if a.CheckFile(f.Name) {
			files = append(files, f)
		}
	}
	sort.Slice(files, func(i, j int) bool {
		return filter.PhpComposerLock(files[i].Name) && !filter.PhpComposerLock(files[j].Name)
	})
	return files
}

// ParseFile Parse the file
func (Analyzer) ParseFile(dirRoot *srt.DirTree, depRoot *srt.DepTree, file *srt.FileData) (deps []*srt.DepTree) {
	deps = []*srt.DepTree{}
	if filter.PhpComposerLock(file.Name) {
		return parseComposerLock(depRoot, file)
	} else if filter.PhpComposer(file.Name) {
		return parseComposer(depRoot, file)
	}
	return deps
}
