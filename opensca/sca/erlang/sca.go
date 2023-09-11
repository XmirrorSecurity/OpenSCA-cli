package erlang

import (
	"context"
	"regexp"

	"github.com/xmirrorsecurity/opensca-cli/opensca/model"
	"github.com/xmirrorsecurity/opensca-cli/opensca/sca/filter"
)

type Sca struct{}

func (sca Sca) Language() model.Language {
	return model.Lan_Erlang
}

func (sca Sca) Filter(relpath string) bool {
	return filter.ErlangRebarLock(relpath)
}

func (sca Sca) Sca(ctx context.Context, parent *model.File, files []*model.File) []*model.DepGraph {
	var deps []*model.DepGraph
	for _, f := range files {
		if sca.Filter(f.Relpath()) {
			deps = append(deps, ParseRebarLock(f))
		}
	}
	return deps
}

func ParseRebarLock(file *model.File) *model.DepGraph {
	reg := regexp.MustCompile(`<<"([\w\d]+)">>\S*?pkg,<<"[\w\d]+">>,<<"([.\d]+)">>`)
	root := &model.DepGraph{Path: file.Relpath()}
	file.ReadLine(func(line string) {
		match := reg.FindStringSubmatch(line)
		root.AppendChild(&model.DepGraph{
			Name:    match[0],
			Version: match[1],
		})
	})
	return root
}
