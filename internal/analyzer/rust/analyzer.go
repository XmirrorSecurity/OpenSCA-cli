package rust

import (
	"opensca/internal/enum/language"
	"opensca/internal/filter"
	"opensca/internal/srt"
)

type Analyzer struct{}

func New() Analyzer {
	return Analyzer{}
}

/**
 * @description: Get language of Analyzer
 * @return {language.Type} language type
 */
func (a Analyzer) GetLanguage() language.Type {
	return language.Rust
}

/**
 * @description: Check if it is a parsable file
 * @param {string} filename file name
 * @return {bool} is a parseable file returns true
 */
func (a Analyzer) CheckFile(filename string) bool {
	return filter.RustCargoLock(filename)
}

/**
 * @description: filters the files that the current parser needs to parse
 * @param {*srt.DirTree} dirRoot directory tree node
 * @param {*srt.DepTree} depRoot Dependency tree node
 * @return {[]*srt.FileData} List of files to parse
 */
func (a Analyzer) FilterFile(dirRoot *srt.DirTree, depRoot *srt.DepTree) []*srt.FileData {
	files := []*srt.FileData{}
	for _, f := range dirRoot.Files {
		if a.CheckFile(f.Name) {
			files = append(files, f)
		}
	}
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
	if filter.RustCargoLock(file.Name) {
		return parseCargoLock(dirRoot, depRoot, file)
	}
	return []*srt.DepTree{}
}
