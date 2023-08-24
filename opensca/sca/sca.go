package sca

import (
	"context"

	"github.com/xmirrorsecurity/opensca-cli/opensca/model"
	"github.com/xmirrorsecurity/opensca-cli/opensca/sca/erlang"
	"github.com/xmirrorsecurity/opensca-cli/opensca/sca/golang"
	"github.com/xmirrorsecurity/opensca-cli/opensca/sca/java"
	"github.com/xmirrorsecurity/opensca-cli/opensca/sca/javascript"
	"github.com/xmirrorsecurity/opensca-cli/opensca/sca/php"
	"github.com/xmirrorsecurity/opensca-cli/opensca/sca/python"
	"github.com/xmirrorsecurity/opensca-cli/opensca/sca/ruby"
	"github.com/xmirrorsecurity/opensca-cli/opensca/sca/rust"
)

type Sca interface {
	Language() model.Language
	Filter(relpath string) bool
	Sca(ctx context.Context, parent *model.File, files []*model.File) []*model.DepGraph
}

var allSca = []Sca{
	java.Sca{},
	python.Sca{},
	javascript.Sca{},
	golang.Sca{},
	ruby.Sca{},
	rust.Sca{},
	erlang.Sca{},
	php.Sca{},
}

func RegisterSca(sca ...Sca) { allSca = sca }

func Filter(relpath string) bool {
	for _, sca := range allSca {
		if sca.Filter(relpath) {
			return true
		}
	}
	return false
}

func Do(ctx context.Context, do func(dep *model.DepGraph)) func(parent *model.File, files []*model.File) {
	return func(parent *model.File, files []*model.File) {
		for _, sca := range allSca {
			for _, dep := range sca.Sca(ctx, parent, files) {
				dep.ForEachNode(func(p, n *model.DepGraph) bool {
					// 补全路径
					if p != nil && len(n.Paths) == 0 {
						n.Paths = make([]string, len(p.Paths))
						copy(n.Paths, p.Paths)
					}
					if n.Name != "" {
						for i := range n.Paths {
							n.Paths[i] += n.Index()
						}
					}
					// 补全语言
					if n.Language == model.Lan_None {
						n.Language = sca.Language()
					}
					// 传递develop
					n.Develop = n.IsDevelop()
					return true
				})
				// 回调
				do(dep)
			}
		}
	}
}
