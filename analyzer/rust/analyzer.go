package rust

import (
	"util/enum/language"
	"util/filter"
	"util/model"
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

// FilterFile filters the files that the current parser needs to parse
func (a Analyzer) FilterFile(dirRoot *model.DirTree, depRoot *model.DepTree) []*model.FileData {
	files := []*model.FileData{}
	for _, f := range dirRoot.Files {
		if a.CheckFile(f.Name) {
			files = append(files, f)
		}
	}
	return files
}

// ParseFile Parse the file
func (a Analyzer) ParseFile(dirRoot *model.DirTree, depRoot *model.DepTree, file *model.FileData) []*model.DepTree {
	if filter.RustCargoLock(file.Name) {
		return parseCargoLock(dirRoot, depRoot, file)
	}
	return []*model.DepTree{}
}
