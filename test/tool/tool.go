package tool

import (
	"strings"

	"github.com/xmirrorsecurity/opensca-cli/opensca/model"
)

func Diff(a, b *model.DepGraph) bool {

	sa := strings.Builder{}
	sb := strings.Builder{}

	a.ForEachNode(func(p, n *model.DepGraph) bool {
		sa.WriteString(n.Index())
		return true
	})

	b.ForEachNode(func(p, n *model.DepGraph) bool {
		sb.WriteString(n.Index())
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
