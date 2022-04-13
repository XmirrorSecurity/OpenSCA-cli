/*
 * @Descripation: ruby解析器
 * @Date: 2021-11-30 14:36:49
 */

package ruby

import (
	"opensca/internal/enum/language"
	"opensca/internal/filter"
	"opensca/internal/srt"
)

type Analyzer struct{}

// New 创建ruby解析器
func New() Analyzer {
	return Analyzer{}
}

// GetLanguage Get language of Analyzer
func (a Analyzer) GetLanguage() language.Type {
	return language.Ruby
}

// CheckFile Check if it is a parsable file
func (a Analyzer) CheckFile(filename string) bool {
	return filter.RubyGemfileLock(filename)
}

// FilterFile filters the files that the current parser needs to parse
func (a Analyzer) FilterFile(dirRoot *srt.DirTree, depRoot *srt.DepTree) []*srt.FileData {
	files := []*srt.FileData{}
	// 筛选需要解析的文件
	for _, f := range dirRoot.Files {
		if a.CheckFile(f.Name) {
			files = append(files, f)
		}
	}
	return files
}

// ParseFile Parse the file
func (a Analyzer) ParseFile(dirRoot *srt.DirTree, depRoot *srt.DepTree, file *srt.FileData) (deps []*srt.DepTree) {
	deps = []*srt.DepTree{}
	if filter.RubyGemfileLock(file.Name) {
		return parseGemfileLock(depRoot, file)
	}
	return deps
}
