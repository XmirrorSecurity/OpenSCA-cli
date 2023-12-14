package rust

import (
	"context"

	"github.com/xmirrorsecurity/opensca-cli/v3/opensca/model"
	"github.com/xmirrorsecurity/opensca-cli/v3/opensca/sca/filter"
)

type Sca struct{}

func (sca Sca) Language() model.Language {
	return model.Lan_Rust
}

func (sca Sca) Filter(relpath string) bool {
	return filter.RustCargoLock(relpath)
}

func (sca Sca) Sca(ctx context.Context, parent *model.File, files []*model.File, call model.ResCallback) {
	for _, f := range files {
		if filter.RustCargoLock(f.Relpath()) {
			root := ParseCargoLock(f)
			if root != nil && len(root.Children) > 0 {
				call(f, root)
			}
		}
	}
}
