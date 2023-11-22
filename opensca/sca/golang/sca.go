package golang

import (
	"context"
	"path/filepath"

	"github.com/xmirrorsecurity/opensca-cli/opensca/model"
	"github.com/xmirrorsecurity/opensca-cli/opensca/sca/filter"
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
	path2dir := func(s string) string { return filepath.Dir(s) }

	// 记录相关文件
	for _, f := range files {
		if filter.GoPkgToml(f.Relpath()) {
			pkgtoml[path2dir(f.Relpath())] = f
		}
		if filter.GoPkgLock(f.Relpath()) {
			pkglock[path2dir(f.Relpath())] = f
		}
		if filter.GoMod(f.Relpath()) {
			gomod[path2dir(f.Relpath())] = f
		}
		if filter.GoSum(f.Relpath()) {
			gosum[path2dir(f.Relpath())] = f
		}
	}

	// 尝试调用 go mod graph
	if len(gomod) > 0 {
		for k, f := range gomod {
			root := GoModGraph(ctx, f)
			if root != nil && len(root.Children) > 0 {
				call(f, root)
				delete(gomod, k)
			}
		}
	}

	for _, f := range gomod {
		mod := ParseGomod(f)
		if sumf, ok := gosum[path2dir(f.Relpath())]; ok {
			sum := ParseGosum(sumf)
			if len(sum.Children) >= len(mod.Children) {
				mod = sum
			}
		}
		call(f, mod)
	}

	for _, f := range pkgtoml {
		pkg := ParseGopkgToml(f)
		if lockf, ok := pkglock[path2dir(f.Relpath())]; ok {
			lock := ParseGopkgLock(lockf)
			if len(lock.Children) >= len(pkg.Children) {
				pkg = lock
			}
		}
		call(f, pkg)
	}

}
