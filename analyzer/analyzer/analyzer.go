package analyzer

import (
	"github.com/xmirrorsecurity/opensca-cli/util/enum/language"
	"github.com/xmirrorsecurity/opensca-cli/util/model"
)

type Analyzer interface {

	// GetLanguage get language of analyzer
	GetLanguage() language.Type

	// CheckFile check parsable file
	CheckFile(filename string) bool

	// ParseFiles parse dependency from file
	ParseFiles(files []*model.FileInfo) []*model.DepTree
}
