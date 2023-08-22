package maven

import (
	"context"

	"github.com/xmirrorsecurity/opensca-cli/opensca/model"
)

type Sca struct{}

func (sca Sca) Filter(relpath string) bool {
	panic("not implemented") // TODO: Implement
}

func (sca Sca) Sca(ctx context.Context, parent model.File, files []model.File) []*model.DepGraph {
	panic("not implemented") // TODO: Implement
}
