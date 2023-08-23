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
	return nil
}
