package ui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/xmirrorsecurity/opensca-cli/cmd/detail"
	"github.com/xmirrorsecurity/opensca-cli/cmd/format"
)

func OpenUI(report format.Report) {

	root := tview.NewTreeNode(report.AppName).SetColor(tcell.ColorGreen)
	depTreeRoot := report.DepDetailGraph
	depTreeRoot.Expand = root

	tree := tview.NewTreeView().SetRoot(root).SetCurrentNode(root)
	tree.SetSelectedFunc(func(node *tview.TreeNode) {
		if len(node.GetChildren()) > 0 {
			node.SetExpanded(!node.IsExpanded())
		}
	})

	depTreeRoot.ForEach(func(n *detail.DepDetailGraph) bool {
		node := n.Expand.(*tview.TreeNode)
		for _, c := range n.Children {
			sub := newTreeNode(c)
			c.Expand = sub
			node.AddChild(sub)
		}
		return true
	})

	if err := tview.NewApplication().SetRoot(tree, true).Run(); err != nil {
		panic(err)
	}
}

func newTreeNode(d *detail.DepDetailGraph) *tview.TreeNode {

	dev := ""
	if d.Develop {
		dev = "<dev>"
	}

	dep := fmt.Sprintf("%s:%s", d.Name, d.Version)
	if d.Vendor != "" {
		dep = fmt.Sprintf("%s:%s", d.Vendor, dep)
	}

	info := fmt.Sprintf("%s%s", dev, dep)

	n := tview.NewTreeNode(info).SetColor(tcell.ColorBlue)

	detail := tview.NewTreeNode("detail")
	// detail.SetExpanded(false)

	detail.AddChild(tview.NewTreeNode(fmt.Sprintf("language:%s", d.Language)))
	if len(d.Paths) > 0 {
		paths := tview.NewTreeNode("paths")
		for _, p := range d.Paths {
			paths.AddChild(tview.NewTreeNode(p))
		}
		detail.AddChild(paths)
	}

	if len(d.Vulnerabilities) > 0 {
		vulns := tview.NewTreeNode("vulns")
		for _, v := range d.Vulnerabilities {
			vuln := tview.NewTreeNode(v.Id)
			vuln.AddChild(tview.NewTreeNode(fmt.Sprintf("name:%s", v.Name)))
			vuln.AddChild(tview.NewTreeNode(fmt.Sprintf("cve:%s", v.Cve)))
			vuln.AddChild(tview.NewTreeNode(fmt.Sprintf("description:%s", v.Description)))
			vuln.AddChild(tview.NewTreeNode(fmt.Sprintf("suggestion:%s", v.Suggestion)))
			vulns.AddChild(vuln)
		}
		detail.AddChild(vulns)
	}
	n.AddChild(detail)

	return n
}
