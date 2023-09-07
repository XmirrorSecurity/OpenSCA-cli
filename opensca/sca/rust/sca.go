package rust

import (
	"context"

	"github.com/xmirrorsecurity/opensca-cli/opensca/model"
	"github.com/xmirrorsecurity/opensca-cli/opensca/sca/filter"
)

type Sca struct{}

func (sca Sca) Language() model.Language {
	return model.Lan_Rust
}

func (sca Sca) Filter(relpath string) bool {
	return filter.RustCargoLock(relpath)
}

func (sca Sca) Sca(ctx context.Context, parent *model.File, files []*model.File) []*model.DepGraph {
	var roots []*model.DepGraph
	for _, f := range files{
		if filter.RustCargoLock(f.Relpath){
			roots = append(roots, ParseCargoLock(f))
		}
	}
	return roots
}
