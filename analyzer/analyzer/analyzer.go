/*
 * @Descripation: Analyzer接口
 * @Date: 2021-11-17 21:26:36
 */

package analyzer

import (
	"util/enum/language"
	"util/model"
)

type Analyzer interface {

	// GetLanguage get language of analyzer
	GetLanguage() language.Type

	// CheckFile check parsable file
	CheckFile(filename string) bool

	// FilterFile filters support files
	FilterFile(dirRoot *model.DirTree, depRoot *model.DepTree) []*model.FileData

	// ParseFile parse dependency from file
	ParseFile(dirRoot *model.DirTree, depRoot *model.DepTree, file *model.FileData) []*model.DepTree
}
