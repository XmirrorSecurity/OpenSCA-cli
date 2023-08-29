package python

import (
	"context"

	"github.com/xmirrorsecurity/opensca-cli/opensca/model"
	"github.com/xmirrorsecurity/opensca-cli/opensca/sca/filter"
)

type Sca struct{}

func (sca Sca) Language() model.Language {
	return model.Lan_Python
}

func (sca Sca) Filter(relpath string) bool {
	return filter.PythonPipfile(relpath) ||
		filter.PythonPipfileLock(relpath) ||
		filter.PythonRequirementsIn(relpath) ||
		filter.PythonRequirementsTxt(relpath) ||
		filter.PythonSetup(relpath)
}

func (sca Sca) Sca(ctx context.Context, parent *model.File, files []*model.File) []*model.DepGraph {
	for _, file := range files {
		if filter.PythonPipfile(file.Relpath) {
			ParsePipfile(file)
		} else if filter.PythonPipfileLock(file.Relpath) {
			ParsePipfileLock(file)
		} else if filter.PythonRequirementsIn(file.Relpath) {
			// TODO
		} else if filter.PythonRequirementsTxt(file.Relpath) {
			// TODO
		} else if filter.PythonSetup(file.Relpath) {
			ParseSetup(file)
		}
	}
	return nil
}
