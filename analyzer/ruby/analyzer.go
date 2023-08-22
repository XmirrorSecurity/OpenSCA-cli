package ruby

import (
	"github.com/xmirrorsecurity/opensca-cli/util/enum/language"
	"github.com/xmirrorsecurity/opensca-cli/util/filter"
	"github.com/xmirrorsecurity/opensca-cli/util/model"
)

type Analyzer struct{}

// New 创建ruby解析器
func New() Analyzer {
	return Analyzer{}
}

// GetLanguage Get language of Analyzer
func (a Analyzer) GetLanguage() language.Type {
	return language.Ruby
}

// CheckFile Check if it is a parsable file
func (a Analyzer) CheckFile(filename string) bool {
	return filter.RubyGemfileLock(filename)
}

// ParseFiles Parse the file
func (a Analyzer) ParseFiles(files []*model.FileInfo) (deps []*model.DepTree) {
	deps = []*model.DepTree{}
	for _, f := range files {
		dep := model.NewDepTree(nil)
		dep.Path = f.Name
		if filter.RubyGemfileLock(f.Name) {
			parseGemfileLock(dep, f)
		}
		deps = append(deps, dep)
	}
	return deps
}
