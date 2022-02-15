/*
 * @Descripation: Analyzer接口
 * @Date: 2021-11-17 21:26:36
 */

package analyzer

import (
	"opensca/internal/enum/language"
	"opensca/internal/srt"
)

type Analyzer interface {

	/**
	 * @description: Get language of Analyzer
	 * @return {language.Type} language type
	 */
	GetLanguage() language.Type

	/**
	 * @description: Check if it is a parsable file
	 * @param {string} filename file name
	 * @return {bool} is a parseable file returns true
	 */
	CheckFile(filename string) bool

	/**
	 * @description: filters the files that the current parser needs to parse
	 * @param {*srt.DirTree} dirRoot directory tree node
	 * @param {*srt.DepTree} depRoot Dependency tree node
	 * @return {[]*srt.FileData} List of files to parse
	 */
	FilterFile(dirRoot *srt.DirTree, depRoot *srt.DepTree) []*srt.FileData

	/**
	 * @description: Parse the file
	 * @param {*srt.DirTree} dirRoot directory tree node
	 * @param {*srt.DepTree} depRoot Dependency tree node
	 * @param {*srt.FileData} file data to parse
	 * @return {[]*srt.DepTree} parsed dependency list
	 */
	ParseFile(dirRoot *srt.DirTree, depRoot *srt.DepTree, file *srt.FileData) []*srt.DepTree
}
