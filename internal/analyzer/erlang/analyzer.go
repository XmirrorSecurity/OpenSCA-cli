package erlang

import (
	"opensca/internal/enum/language"
	"opensca/internal/filter"
	"opensca/internal/srt"
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
func (a Analyzer) FilterFile(dirRoot *srt.DirTree, depRoot *srt.DepTree) []*srt.FileData {
	files := []*srt.FileData{}
	for _, f := range dirRoot.Files {
		if a.CheckFile(f.Name) {
			files = append(files, f)
		}
	}
	return files
}

// ParseFile parse dependency from file
func (a Analyzer) ParseFile(dirRoot *srt.DirTree, depRoot *srt.DepTree, file *srt.FileData) []*srt.DepTree {
	deps := []*srt.DepTree{}
	if filter.ErlangRebarLock(file.Name) {
		deps = parseRebarLock(depRoot, file)
	}
	return deps
}
