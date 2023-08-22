package golang

import (
	"github.com/xmirrorsecurity/opensca-cli/util/enum/language"
	"github.com/xmirrorsecurity/opensca-cli/util/filter"
	"github.com/xmirrorsecurity/opensca-cli/util/model"
)

// golang Analyzer
type Analyzer struct{}

// New create golang Analyzer
func New() Analyzer {
	return Analyzer{}
}

// GetLanguage Get language of Analyzer
func (Analyzer) GetLanguage() language.Type {
	return language.Golang
}

// CheckFile Check if it is a parsable file
func (Analyzer) CheckFile(filename string) bool {
	return filter.GoMod(filename) || filter.GoSum(filename)
}

// ParseFiles Parse the file
func (Analyzer) ParseFiles(files []*model.FileInfo) []*model.DepTree {
	deps := []*model.DepTree{}
	for _, f := range files {
		dep := model.NewDepTree(nil)
		dep.Path = f.Name
		if filter.GoMod(f.Name) {
			parseGomod(dep, f)
		} else if filter.GoSum(f.Name) {
			parseGosum(dep, f)
		}
		deps = append(deps, dep)
	}
	return deps
}
