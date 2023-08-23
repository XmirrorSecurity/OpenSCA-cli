package java

import (
	"context"
	"strings"

	"github.com/xmirrorsecurity/opensca-cli/opensca/model"
	"github.com/xmirrorsecurity/opensca-cli/opensca/sca/filter"
)

type Sca struct{}

func (sca Sca) Filter(relpath string) bool {
	return filter.JavaPom(relpath)
}

func (sca Sca) Sca(ctx context.Context, parent *model.File, files []*model.File) []*model.DepGraph {

	if strings.Contains(parent.Relpath, ".jar") {
		// TODO: 仅解析pom本身信息
		return nil
	}

	// 调用mvn解析
	deps := MvnTree(parent.Abspath)
	if len(deps) > 0 {
		return deps
	}

	// 模拟maven构建
	pomfiles := []*model.File{}
	for _, file := range files {
		if sca.Filter(parent.Relpath) {
			pomfiles = append(pomfiles, file)
		}
	}
	return Simulate(pomfiles)
}
