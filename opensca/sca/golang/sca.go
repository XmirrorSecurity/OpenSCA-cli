package golang

import (
	"context"
	"path/filepath"

	"github.com/xmirrorsecurity/opensca-cli/v3/cmd/config"
	"github.com/xmirrorsecurity/opensca-cli/v3/opensca/model"
	"github.com/xmirrorsecurity/opensca-cli/v3/opensca/sca/filter"
)

type Sca struct{}

func (sca Sca) Language() model.Language {
	return model.Lan_Golang
}

func (sca Sca) Filter(relpath string) bool {
	return filter.GoMod(relpath) || filter.GoSum(relpath) || filter.GoPkgToml(relpath) || filter.GoPkgLock(relpath)
}

func (sca Sca) Sca(ctx context.Context, parent *model.File, files []*model.File, call model.ResCallback) {

	// map[dir]*File
	gomod := map[string]*model.File{}
	gosum := map[string]*model.File{}
	pkglock := map[string]*model.File{}
	pkgtoml := map[string]*model.File{}

	// 记录相关文件
	for _, f := range files {
		dir := filepath.Dir(f.Relpath())
		if filter.GoPkgToml(f.Relpath()) {
			pkgtoml[dir] = f
		}
		if filter.GoPkgLock(f.Relpath()) {
			pkglock[dir] = f
		}
		if filter.GoMod(f.Relpath()) {
			gomod[dir] = f
		}
		if filter.GoSum(f.Relpath()) {
			gosum[dir] = f
		}
	}

	// 尝试调用 go mod graph
	if config.Conf().Optional.Dynamic {
		for dir, f := range gomod {
			graph := GoModGraph(ctx, f)
			if graph != nil && len(graph.Children) > 0 {
				call(f, graph)
				delete(gomod, dir)
				delete(gosum, dir)
			}
		}
	}

	// 静态解析go.sum
	for dir, f := range gosum {
		sum := ParseGosum(f)
		call(f, sum)
		delete(gomod, dir)
	}

	// 静态解析go.mod
	for _, f := range gomod {
		mod := ParseGomod(f)
		call(f, mod)
	}

	// 静态解析gopkg.lock
	for dir, f := range pkglock {
		lock := ParseGopkgLock(f)
		call(f, lock)
		delete(pkgtoml, dir)
	}

	// 静态解析gopkg.toml
	for _, f := range pkgtoml {
		pkg := ParseGopkgToml(f)
		call(f, pkg)
	}

}
