package rust

import (
	"github.com/xmirrorsecurity/opensca-cli/util/enum/language"
	"github.com/xmirrorsecurity/opensca-cli/util/filter"
	"github.com/xmirrorsecurity/opensca-cli/util/model"
)

type Analyzer struct{}

func New() Analyzer {
	return Analyzer{}
}

// GetLanguage Get language of Analyzer
func (a Analyzer) GetLanguage() language.Type {
	return language.Rust
}

// CheckFile Check if it is a parsable file
func (a Analyzer) CheckFile(filename string) bool {
	return filter.RustCargoLock(filename)
}

// ParseFiles Parse the file
func (a Analyzer) ParseFiles(files []*model.FileInfo) []*model.DepTree {
	deps := []*model.DepTree{}
	for _, f := range files {
		dep := model.NewDepTree(nil)
		dep.Path = f.Name
		if filter.RustCargoLock(f.Name) {
			parseCargoLock(dep, f)
		}
		deps = append(deps, dep)
	}
	return deps
}
