package erlang

import (
	"context"

	"github.com/xmirrorsecurity/opensca-cli/opensca/model"
	"github.com/xmirrorsecurity/opensca-cli/opensca/sca/filter"
)

type Sca struct{}

func (sca Sca) Filter(relpath string) bool {
	return filter.ErlangRebarLock(relpath)
}

func (sca Sca) Sca(ctx context.Context, parent *model.File, files []*model.File) []*model.DepGraph {
	return nil
}
