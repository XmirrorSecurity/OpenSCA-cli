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

/**
 * @description: 创建php解析器
 * @return {php.Analyzer} php解析器
 */
func New() Analyzer {
	return Analyzer{}
}

/**
 * @description: Get language of Analyzer
 * @return {language.Type} language type
 */
func (Analyzer) GetLanguage() language.Type {
	return language.Php
}

/**
 * @description: Check if it is a parsable file
 * @param {string} filename file name
 * @return {bool} is a parseable file returns true
 */
func (Analyzer) CheckFile(filename string) bool {
	return filter.PhpComposerLock(filename) || filter.PhpComposer(filename)
}

/**
 * @description: filters the files that the current parser needs to parse
 * @param {*srt.DirTree} dirRoot directory tree node
 * @param {*srt.DepTree} depRoot Dependency tree node
 * @return {[]*srt.FileData} List of files to parse
 */
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

/**
 * @description: Parse the file
 * @param {*srt.DirTree} dirRoot directory tree node
 * @param {*srt.DepTree} depRoot Dependency tree node
 * @param {*srt.FileData} file data to parse
 * @return {[]*srt.DepTree} parsed dependency list
 */
func (Analyzer) ParseFile(dirRoot *srt.DirTree, depRoot *srt.DepTree, file *srt.FileData) (deps []*srt.DepTree) {
	deps = []*srt.DepTree{}
	if filter.PhpComposerLock(file.Name) {
		return parseComposerLock(depRoot, file)
	} else if filter.PhpComposer(file.Name) {
		return parseComposer(depRoot, file)
	}
	return deps
}
