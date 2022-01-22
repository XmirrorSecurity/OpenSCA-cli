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
 * @description: 获取当前Analyzer的语言
 * @return {language.Type} 语言
 */
func (Analyzer) GetLanguage() language.Type {
	return language.Php
}

/**
 * @description: 检测是否是可解析的文件
 * @param {string} filename 文件名
 * @return {bool} 是可解析的文件返回true
 */
func (Analyzer) CheckFile(filename string) bool {
	return filter.PhpComposerLock(filename) || filter.PhpComposer(filename)
}

/**
 * @descriptsrt筛选当前解析器需要解析的文件
 * @param {*modsrtrTree} dirRoot 目录树节点
 * @param {*srt.DepTree} depRoot 依赖树节点
 * @return {[]*srt.FileData} 需要解析的文件列表
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
 * @descriptsrt解析文件
 * @param {*srt.DirTree} dirRoot 目录树节点
 * @param {*modsrtpTree} depRoot 依赖树节点
 * @param {*srt.FileData} file 文件信息
 * @return {[]*srt.DepTree} 解析出的依赖列表
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
