package java

import (
	"context"
	"io"
	"strings"

	"github.com/xmirrorsecurity/opensca-cli/opensca/model"
	"github.com/xmirrorsecurity/opensca-cli/opensca/sca/filter"
)

type Sca struct{}

func (sca Sca) Language() model.Language {
	return model.Lan_Java
}

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
	poms := []*Pom{}
	for _, file := range files {
		if sca.Filter(file.Relpath) {
			file.OpenReader(func(reader io.Reader) {
				poms = append(poms, ReadPom(reader))
			})
		}
	}
	deps = append(deps, ParsePoms(poms)...)
	return deps
}

type MvnRepo struct {
	Url      string `json:"url" xml:"url"`
	Username string
	Password string
}

var defaultRepo []MvnRepo

func RegisterRepo(repos ...MvnRepo) {
	defaultRepo = append(repos, repos...)
}
