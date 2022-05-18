package erlang

import (
	"util/enum/language"
	"util/filter"
	"util/model"
)

type Analyzer struct{}

func New() Analyzer {
	return Analyzer{}
}

// GetLanguage get language of analyzer
func (a Analyzer) GetLanguage() language.Type {
	return language.Erlang
}

// CheckFile check parsable file
func (a Analyzer) CheckFile(filename string) bool {
	return filter.ErlangRebarLock(filename)
}

// ParseFiles parse dependency from file
func (a Analyzer) ParseFiles(files []*model.FileInfo) []*model.DepTree {
	deps := []*model.DepTree{}
	for _, f := range files {
		dep := model.NewDepTree(nil)
		dep.Path = f.Name
		if filter.ErlangRebarLock(f.Name) {
			parseRebarLock(dep, f)
		}
		deps = append(deps, dep)
	}
	return deps
}
