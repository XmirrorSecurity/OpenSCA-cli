package erlang

import (
	"context"
	"regexp"

	"github.com/xmirrorsecurity/opensca-cli/v3/opensca/model"
	"github.com/xmirrorsecurity/opensca-cli/v3/opensca/sca/filter"
)

type Sca struct{}

func (sca Sca) Language() model.Language {
	return model.Lan_Erlang
}

func (sca Sca) Filter(relpath string) bool {
	return filter.ErlangRebarLock(relpath)
}

func (sca Sca) Sca(ctx context.Context, parent *model.File, files []*model.File, call model.ResCallback) {
	for _, f := range files {
		if sca.Filter(f.Relpath()) {
			call(f, ParseRebarLock(f))
		}
	}
}

func ParseRebarLock(file *model.File) *model.DepGraph {
	reg := regexp.MustCompile(`<<"([\w\d]+)">>\S*?pkg,<<"[\w\d]+">>,<<"([.\d]+)">>`)
	root := &model.DepGraph{Path: file.Relpath()}
	file.ReadLine(func(line string) {
		match := reg.FindStringSubmatch(line)
		if len(match) > 2 {
			root.AppendChild(&model.DepGraph{
				Name:    match[1],
				Version: match[2],
			})
		}
	})
	return root
}
