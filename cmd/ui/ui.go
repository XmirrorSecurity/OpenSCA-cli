package ui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/xmirrorsecurity/opensca-cli/cmd/detail"
	"github.com/xmirrorsecurity/opensca-cli/cmd/format"
)

var (
	colorVul    = tcell.ColorPink
	colorPath   = tcell.ColorBlue
	colorDep    = tcell.ColorGreen
	colorDevDep = tcell.ColorGrey
	colorVulDep = tcell.ColorRed
)

func OpenUI(report format.Report) {

	root := tview.NewTreeNode(report.AppName).SetColor(tcell.ColorBlue)
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

	// 组件信息文本
	dev := ""
	if d.Develop {
		dev = "<dev>"
	}
	dep := fmt.Sprintf("%s:%s", d.Name, d.Version)
	if d.Vendor != "" {
		dep = fmt.Sprintf("%s:%s", d.Vendor, dep)
	}
	info := fmt.Sprintf("%s%s<%s>", dev, dep, d.Language)

	// 当前节点
	n := tview.NewTreeNode(info)

	// 没有子依赖则不展开
	if len(d.Children) == 0 {
		n.SetExpanded(false)
	}

	n.SetColor(colorDep)
	if len(d.Vulnerabilities) > 0 {
		n.SetColor(colorVulDep)
	}
	if d.Develop {
		n.SetColor(colorDevDep)
	}

	// 路径
	if len(d.Paths) > 0 {
		paths := tview.NewTreeNode("paths")
		paths.SetColor(colorPath)
		paths.SetExpanded(!n.IsExpanded())
		for _, p := range d.Paths {
			paths.AddChild(tview.NewTreeNode(p).SetColor(colorPath))
		}
		n.AddChild(paths)
	}

	// 漏洞
	if len(d.Vulnerabilities) > 0 {
		vulns := tview.NewTreeNode("vulns")
		vulns.SetColor(colorVul)
		vulns.SetExpanded(!n.IsExpanded())
		for _, v := range d.Vulnerabilities {
			// 漏洞=>详细字段
			vuln := tview.NewTreeNode(v.Id).SetColor(colorVul)
			vuln.AddChild(tview.NewTreeNode(fmt.Sprintf("name:%s", v.Name)).SetColor(colorVul))
			vuln.AddChild(tview.NewTreeNode(fmt.Sprintf("cve:%s", v.Cve)).SetColor(colorVul))
			vuln.AddChild(tview.NewTreeNode(fmt.Sprintf("description:%s", v.Description)).SetColor(colorVul))
			vuln.AddChild(tview.NewTreeNode(fmt.Sprintf("suggestion:%s", v.Suggestion)).SetColor(colorVul))
			vulns.AddChild(vuln)
		}
		n.AddChild(vulns)
	}

	return n
}
