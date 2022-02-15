/*
 * @Descripation: javascript解析器
 * @Date: 2021-11-25 19:59:35
 */

package javascript

import (
	"opensca/internal/enum/language"
	"opensca/internal/filter"
	"opensca/internal/srt"
	"sort"
)

type Analyzer struct{}

/**
 * @description: 创建javascript解析器
 * @return {javascript.Analyzer} javascript解析器
 */
func New() Analyzer {
	return Analyzer{}
}

/**
 * @description: Get language of Analyzer
 * @return {language.Type} language type
 */
func (a Analyzer) GetLanguage() language.Type {
	return language.JavaScript
}

/**
 * @description: Check if it is a parsable file
 * @param {string} filename file name
 * @return {bool} is a parseable file returns true
 */
func (a Analyzer) CheckFile(filename string) bool {
	return filter.JavaScriptPackageLock(filename) ||
		filter.JavaScriptPackage(filename) ||
		filter.JavaScriptYarnLock(filename)
}

/**
 * @description: filters the files that the current parser needs to parse
 * @param {*srt.DirTree} dirRoot directory tree node
 * @param {*srt.DepTree} depRoot Dependency tree node
 * @return {[]*srt.FileData} List of files to parse
 */
func (a Analyzer) FilterFile(dirRoot *srt.DirTree, depRoot *srt.DepTree) (files []*srt.FileData) {
	files = []*srt.FileData{}
	// 标记是否存在lock文件
	lock := false
	// 筛选需要解析的文件
	for _, f := range dirRoot.Files {
		if a.CheckFile(f.Name) {
			files = append(files, f)
			if filter.JavaScriptPackageLock(f.Name) {
				lock = true
			}
		}
	}
	// 存在package-lock.json文件则不解析package.json文件
	if lock {
		for i := 0; i < len(files); {
			if filter.JavaScriptPackage(files[i].Name) {
				files = append(files[:i], files[i+1:]...)
			} else {
				i++
			}
		}
	}
	// 文件排序
	sort.Slice(files, func(i, j int) bool {
		// 优先解析package-lock.json文件
		return filter.JavaScriptPackageLock(files[i].Name) && !filter.JavaScriptPackageLock(files[j].Name)
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
func (a Analyzer) ParseFile(dirRoot *srt.DirTree, depRoot *srt.DepTree, file *srt.FileData) []*srt.DepTree {
	deps := []*srt.DepTree{}
	if filter.JavaScriptPackageLock(file.Name) {
		return parsePackageLock(depRoot, file)
	} else if filter.JavaScriptPackage(file.Name) {
		return parsePackage(depRoot, file)
	} else if filter.JavaScriptYarnLock(file.Name) {
		return parseYarnLock(depRoot, file)
	}
	return deps
}
