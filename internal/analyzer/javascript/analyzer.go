/*
 * @Descripation: javascript解析器
 * @Date: 2021-11-25 19:59:35
 */

package javascript

import (
	"opensca/internal/enum/language"
	"opensca/internal/filter"
	"opensca/internal/srt"
)

type Analyzer struct{}

// New 创建javascript解析器
func New() Analyzer {
	return Analyzer{}
}

// GetLanguage Get language of Analyzer
func (a Analyzer) GetLanguage() language.Type {
	return language.JavaScript
}

// CheckFile Check if it is a parsable file
func (a Analyzer) CheckFile(filename string) bool {
	return filter.JavaScriptPackageLock(filename) ||
		filter.JavaScriptPackage(filename) ||
		filter.JavaScriptYarnLock(filename)
}

// FilterFile filters the files that the current parser needs to parse
func (a Analyzer) FilterFile(dirRoot *srt.DirTree, depRoot *srt.DepTree) (files []*srt.FileData) {
	files = []*srt.FileData{}
	// 标记是否存在lock文件
	lock := false
	// 筛选需要解析的文件
	for _, f := range dirRoot.Files {
		if a.CheckFile(f.Name) {
			files = append(files, f)
			if filter.JavaScriptPackageLock(f.Name) || filter.JavaScriptYarnLock(f.Name) {
				lock = true
			}
		}
	}
	// 存在 yarn.lock 或 package-lock.json 文件则不解析package.json文件
	if lock {
		for i := 0; i < len(files); {
			if filter.JavaScriptPackage(files[i].Name) {
				files = append(files[:i], files[i+1:]...)
			} else {
				i++
			}
		}
	}
	return files
}

// ParseFile Parse the file
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
