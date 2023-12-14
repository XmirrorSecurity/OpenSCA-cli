package ruby

import (
	"context"

	"github.com/xmirrorsecurity/opensca-cli/v3/opensca/model"
	"github.com/xmirrorsecurity/opensca-cli/v3/opensca/sca/filter"
)

type Sca struct{}

func (sca Sca) Language() model.Language {
	return model.Lan_Ruby
}

func (sca Sca) Filter(relpath string) bool {
	return filter.RubyGemfileLock(relpath)
}

func (sca Sca) Sca(ctx context.Context, parent *model.File, files []*model.File, call model.ResCallback) {
	for _, file := range files {
		if filter.RubyGemfileLock(file.Relpath()) {
			call(file, ParseGemfileLock(file))
		}
	}
}
