/*
 * @Description: Analyzer接口
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

	// ParseFiles parse dependency from file
	ParseFiles(files []*model.FileInfo) []*model.DepTree
}
