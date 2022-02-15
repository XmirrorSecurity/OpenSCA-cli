/*
 * @Description: golang analyzer
 * @Date: 2022-02-10 16:08:00
 */

package golang

import (
	"opensca/internal/enum/language"
	"opensca/internal/filter"
	"opensca/internal/srt"
	"sort"
)

// golang Analyzer
type Analyzer struct{}

/**
 * @description: create golang Analyzer
 * @return {golang.Analyzer} golang Analyzer
 */
func New() Analyzer {
	return Analyzer{}
}

/**
 * @description: Get language of Analyzer
 * @return {language.Type} language type
 */
func (Analyzer) GetLanguage() language.Type {
	return language.Golang
}

/**
 * @description: Check if it is a parsable file
 * @param {string} filename file name
 * @return {bool} is a parseable file returns true
 */
func (Analyzer) CheckFile(filename string) bool {
	return filter.GoMod(filename) || filter.GoSum(filename)
}

/**
 * @description: filters the files that the current parser needs to parse
 * @param {*srt.DirTree} dirRoot directory tree node
 * @param {*srt.DepTree} depRoot Dependency tree node
 * @return {[]*srt.FileData} List of files to parse
 */
func (a Analyzer) FilterFile(dirRoot *srt.DirTree, depRoot *srt.DepTree) []*srt.FileData {
	files := []*srt.FileData{}
	for _, file := range dirRoot.Files {
		if a.CheckFile(file.Name) {
			files = append(files, file)
		}
	}
	sort.Slice(files, func(i, j int) bool {
		return filter.GoSum(files[i].Name) && !filter.GoSum(files[j].Name)
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
func (Analyzer) ParseFile(dirRoot *srt.DirTree, depRoot *srt.DepTree, file *srt.FileData) []*srt.DepTree {
	if filter.GoMod(file.Name) {
		return parseGomod(depRoot, file)
	} else if filter.GoSum(file.Name) {
		return parseGosum(depRoot, file)
	}
	return []*srt.DepTree{}
}
