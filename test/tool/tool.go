package tool

import (
	"github.com/xmirrorsecurity/opensca-cli/opensca/model"
)

func Diff(a, b *model.DepGraph) bool {
	clear := func(p, n *model.DepGraph) bool {
		n.Path = ""
		n.Language = model.Lan_None
		return true
	}
	a.ForEachNode(clear)
	b.ForEachNode(clear)
	return a.Tree(false, true) != b.Tree(false, true)
}

func Dep(vendor, name, version string, children ...*model.DepGraph) *model.DepGraph {
	root := &model.DepGraph{
		Vendor:  vendor,
		Name:    name,
		Version: version,
	}
	for _, c := range children {
		root.AppendChild(c)
	}
	return root
}

func DevDep(vendor, name, version string, children ...*model.DepGraph) *model.DepGraph {
	root := Dep(vendor, name, version, children...)
	root.Develop = true
	return root
}
