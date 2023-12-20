package groovy

import (
	"context"

	"github.com/xmirrorsecurity/opensca-cli/v3/opensca/model"
	"github.com/xmirrorsecurity/opensca-cli/v3/opensca/sca/filter"
)

type Sca struct{}

func (sca Sca) Language() model.Language {
	return model.Lan_Java
}

func (sca Sca) Filter(relpath string) bool {
	return filter.GroovyGradle(relpath) || filter.GroovyFile(relpath)
}

func (sca Sca) Sca(ctx context.Context, parent *model.File, files []*model.File, call model.ResCallback) {

	roots := GradleTree(ctx, parent)
	if len(roots) == 0 {
		roots = ParseGradle(ctx, files)
	}
	if len(roots) > 0 {
		call(parent, roots...)
	}

	for _, f := range files {
		if filter.GroovyFile(f.Relpath()) {
			call(f, ParseGroovy(f))
		}
	}
}
