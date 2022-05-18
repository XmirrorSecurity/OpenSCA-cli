package groovy

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
	return language.Groovy
}

// CheckFile check parsable file
func (a Analyzer) CheckFile(filename string) bool {
	return false
	// groovy 文件无法解析依赖层级，暂不处理
	// return filter.GroovyFile(filename)
}

// ParseFiles parse dependency from file
func (a Analyzer) ParseFiles(files []*model.FileInfo) []*model.DepTree {
	deps := []*model.DepTree{}
	for _, f := range files {
		dep := model.NewDepTree(nil)
		dep.Path = f.Name
		if filter.GroovyFile(f.Name) {
			parseGroovyFile(dep, f)
		}
		deps = append(deps, dep)
	}
	return deps
}
