package ruby

import (
	"context"

	"github.com/xmirrorsecurity/opensca-cli/opensca/model"
	"github.com/xmirrorsecurity/opensca-cli/opensca/sca/filter"
)

type Sca struct{}

func (sca Sca) Language() model.Language {
	return model.Lan_Ruby
}

func (sca Sca) Filter(relpath string) bool {
	return filter.RubyGemfileLock(relpath)
}

func (sca Sca) Sca(ctx context.Context, parent *model.File, files []*model.File) []*model.DepGraph {
	var root []*model.DepGraph
	for _, file := range files {
		if sca.Filter(file.Relpath) {
			root = append(root, ParseGemfileLock(file))
		}
	}
	return root
}
