package tool

import (
	"strings"

	"github.com/xmirrorsecurity/opensca-cli/opensca/model"
)

func Diff(a, b *model.DepGraph) bool {

	sa := strings.Builder{}
	sb := strings.Builder{}

	key := func(d *model.DepGraph) string {
		if d.Develop {
			return "dev:" + d.Index()
		}
		return d.Index()
	}

	a.ForEachNode(func(p, n *model.DepGraph) bool {
		sa.WriteString(key(n))
		return true
	})

	b.ForEachNode(func(p, n *model.DepGraph) bool {
		sb.WriteString(key(n))
		return true
	})

	return sa.String() != sb.String()
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
