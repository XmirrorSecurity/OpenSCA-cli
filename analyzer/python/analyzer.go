package python

import (
	"util/enum/language"
	"util/filter"
	"util/model"
)

type Analyzer struct {
}

func New() Analyzer {
	return Analyzer{}
}

// GetLanguage get language of analyzer
func (Analyzer) GetLanguage() language.Type {
	return language.Python
}

// CheckFile check parsable file
func (Analyzer) CheckFile(filename string) bool {
	return filter.PythonSetup(filename) ||
		filter.PythonPipfile(filename) ||
		filter.PythonPipfileLock(filename) ||
		filter.PythonRequirementsTxt(filename) ||
		filter.PythonRequirementsIn(filename)
}

// ParseFiles parse dependency from file
func (Analyzer) ParseFiles(files []*model.FileInfo) []*model.DepTree {
	deps := []*model.DepTree{}
	for _, f := range files {
		dep := model.NewDepTree(nil)
		dep.Path = f.Name
		if filter.PythonSetup(f.Name) {
			parseSetup(dep, f)
		} else if filter.PythonPipfile(f.Name) {
			parsePipfile(dep, f)
		} else if filter.PythonPipfileLock(f.Name) {
			parsePipfileLock(dep, f)
		} else if filter.PythonRequirementsTxt(f.Name) || filter.PythonRequirementsIn(f.Name) {
			parseRequirementsin(dep, f)
		}
		deps = append(deps, dep)
	}
	return deps
}
