package groovy

import (
	"context"

	"github.com/xmirrorsecurity/opensca-cli/opensca/model"
	"github.com/xmirrorsecurity/opensca-cli/opensca/sca/filter"
)

type Sca struct{}

func (sca Sca) Language() model.Language {
	return model.Lan_Java
}

func (sca Sca) Filter(relpath string) bool {
	return filter.GroovyGradle(relpath) || filter.GroovyFile(relpath)
}

func (sca Sca) Sca(ctx context.Context, parent *model.File, files []*model.File) []*model.DepGraph {
	roots := GradleTree(parent.Abspath())
	if len(roots) == 0 {
		// TODO
		roots = ParseGradle(files)
	}
	return roots
}
