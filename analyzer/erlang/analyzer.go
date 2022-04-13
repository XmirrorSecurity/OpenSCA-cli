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

// FilterFile filters support files
func (a Analyzer) FilterFile(dirRoot *model.DirTree, depRoot *model.DepTree) []*model.FileData {
	files := []*model.FileData{}
	for _, f := range dirRoot.Files {
		if a.CheckFile(f.Name) {
			files = append(files, f)
		}
	}
	return files
}

// ParseFile parse dependency from file
func (a Analyzer) ParseFile(dirRoot *model.DirTree, depRoot *model.DepTree, file *model.FileData) []*model.DepTree {
	deps := []*model.DepTree{}
	if filter.ErlangRebarLock(file.Name) {
		deps = parseRebarLock(depRoot, file)
	}
	return deps
}
