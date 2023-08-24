package javascript

import (
	"context"

	"github.com/xmirrorsecurity/opensca-cli/opensca/model"
	"github.com/xmirrorsecurity/opensca-cli/opensca/sca/filter"
)

type Sca struct{}

func (sca Sca) Language() model.Language {
	return model.Lan_JavaScript
}

func (sca Sca) Filter(relpath string) bool {
	return filter.JavaScriptPackageJson(relpath) ||
		filter.JavaScriptPackageLock(relpath) ||
		filter.JavaScriptYarnLock(relpath)
}

func (sca Sca) Sca(ctx context.Context, parent *model.File, files []*model.File) []*model.DepGraph {
	deps := ParseNpm(files)
	return deps
}
