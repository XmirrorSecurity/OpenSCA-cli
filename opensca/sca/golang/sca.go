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

func (sca Sca) Sca(ctx context.Context, parent *model.File, files []*model.File) []*model.DepGraph {

	gomod := false
	for _, file := range files {
		if filter.GoMod(file.Relpath) {
			gomod = true
			break
		}
	}
	if gomod {
		if root := GoModGraph(ctx, parent.Abspath); root != nil {
			return []*model.DepGraph{root}
		}
	}

	gosum := map[string]*model.File{}
	pkglock := map[string]*model.File{}

	path2dir := func(s string) string { return filepath.Dir(s) }

	for _, f := range files {
		if filter.GoPkgLock(f.Relpath) {
			pkglock[path2dir(f.Relpath)] = f
		}
		if filter.GoSum(f.Relpath) {
			gosum[path2dir(f.Relpath)] = f
		}
	}

	var roots []*model.DepGraph
	for _, f := range files {

		if filter.GoMod(f.Relpath) {
			mod := ParseGomod(f)
			if sumf, ok := gosum[path2dir(f.Relpath)]; ok {
				sum := ParseGosum(sumf)
				if len(sum.Children) > len(mod.Children) {
					mod.Children = sum.Children
				}
			}
			roots = append(roots, mod)
		}

		if filter.GoPkgToml(f.Relpath) {
			pkg := ParseGopkgToml(f)
			if lockf, ok := pkglock[path2dir(f.Relpath)]; ok {
				lock := ParseGopkgLock(lockf)
				if len(lock.Children) > len(pkg.Children) {
					pkg.Children = lock.Children
				}
			}
			roots = append(roots, pkg)
		}

	}

	return roots
}
