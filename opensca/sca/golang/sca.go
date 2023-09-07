package golang

import (
	"context"
	"path/filepath"

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

	gosum := map[string]*model.File{}

	for _, f := range files {
		if filter.GoSum(f.Relpath) {
			gosum[filepath.Dir(f.Relpath)] = f
		}
	}

	var roots []*model.DepGraph
	for _, f := range files {
		if filter.GoMod(f.Relpath) {
			mod := ParseGomod(f)
			if sumf, ok := gosum[filepath.Dir(f.Relpath)]; ok {
				sum := ParseGosum(sumf)
				if len(sum.Children) > len(mod.Children) {
					mod.Children = sum.Children
				}
			}
			roots = append(roots, mod)
		}
	}

	return roots
}
