package golang

import (
	"context"

	"github.com/xmirrorsecurity/opensca-cli/opensca/model"
	"github.com/xmirrorsecurity/opensca-cli/opensca/sca/filter"
)

type Sca struct{}

func (sca Sca) Language() model.Language {
	return model.Lan_Golang
}

func (sca Sca) Filter(relpath string) bool {
	return filter.GoMod(relpath) || filter.GoSum(relpath)
}

func (sca Sca) Sca(ctx context.Context, parent *model.File, files []*model.File) []*model.DepGraph {
	var roots []*model.DepGraph
	for _, f := range files {
		if filter.GoMod(f.Relpath) {
			roots = append(roots, ParseGomod(f))
		}
		if filter.GoSum(f.Relpath) {
			roots = append(roots, ParseGosum(f))
		}
	}
	return roots
}
