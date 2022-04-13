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

	// GetLanguage get language of analyzer
	GetLanguage() language.Type

	// CheckFile check parsable file
	CheckFile(filename string) bool

	// FilterFile filters support files
	FilterFile(dirRoot *srt.DirTree, depRoot *srt.DepTree) []*srt.FileData

	// ParseFile parse dependency from file
	ParseFile(dirRoot *srt.DirTree, depRoot *srt.DepTree, file *srt.FileData) []*srt.DepTree
}
