package ui

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/xmirrorsecurity/opensca-cli/cmd/detail"
	"github.com/xmirrorsecurity/opensca-cli/cmd/format"
	"github.com/xmirrorsecurity/opensca-cli/opensca/logs"
)

var (
	colorVul     = tcell.ColorPurple
	colorPath    = tcell.ColorBlue
	colorDep     = tcell.ColorGreen
	colorDevDep  = tcell.ColorGrey
	colorVulDep  = tcell.ColorRed
	colorLicense = tcell.ColorYellow

	colorLogMap = map[string]string{
		"[TRACE]": "[grey]",
		"[DEBUG]": "[green]",
		"[INFO]":  "[blue]",
		"[WARN]":  "[yellow]",
		"[ERROR]": "[red]",
	}
)

func OpenUI(report format.Report) {

	flex := tview.NewFlex()

	tree := DepTree(report)
	tree.SetBorder(true).SetTitle(" dependency tree ")
	log := TaskLog()
	log.SetBorder(true).SetTitle(" log ")
	info := TaskInfo(report)
	info.SetBorder(true).SetTitle(" info ")
	help := UIHelp()
	help.SetBorder(true).SetTitle(" help ")

	flex.SetDirection(tview.FlexRow).
		AddItem(tview.NewFlex().
			AddItem(tree, 0, 1, true).
			AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(info, 4, 1, false).
				AddItem(log, 0, 1, false),
				0, 1, false),
			0, 1, true).
		AddItem(help, 3, 1, false)

	app := tview.NewApplication()
	switchView := func() {
		if app.GetFocus() == tree {
			app.SetFocus(log)
		} else {
			app.SetFocus(tree)
		}
	}

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {

		switch event.Rune() {
		case 'h', 'l':
			switchView()
		case 'q':
			app.Stop()
		case ']':
			tree.GetCurrentNode().ExpandAll()
		case '[':
			tree.GetCurrentNode().CollapseAll()
		case '}':
			tree.GetRoot().ExpandAll()
		case '{':
			tree.GetRoot().CollapseAll()
		}

		switch event.Key() {
		case tcell.KeyLeft, tcell.KeyRight:
			switchView()
		}

		return event
	})

	if err := app.SetRoot(flex, true).Run(); err != nil {
		logs.Error(err)
	}
}

func TaskInfo(report format.Report) *tview.TextView {
	info := tview.NewTextView().
		SetText(format.Statis(report))
	info.SetTextColor(tcell.ColorBlue)
	return info
}

func UIHelp() *tview.TextView {
	tips := []string{"j:down", "k:up", "space:expand/collapse", "h/l:switch", "g:top", "G:bottom", "crtl+c/q:quit", "{:collapse_all", "}:expand_all", "[:collapse_node", "]:expand_node"}
	text := tview.NewTextView().SetText(strings.Join(tips, " | "))
	return text
}

func TaskLog() *tview.TextView {
	log := tview.NewTextView().SetDynamicColors(true).ScrollToEnd()
	if logs.LogFilePath == "" {
		log.SetText("log file not found")
		return log
	}
	f, err := os.Open(logs.LogFilePath)
	if err != nil {
		log.SetText(fmt.Sprintf("read log file err:\n%s", err))
		return log
	}
	lines := []string{}
	lineNum := 0
	scanner := bufio.NewScanner(f)
	color := "white"
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		if i := strings.Index(line, " "); i != -1 {
			if c, ok := colorLogMap[line[:i]]; ok {
				color = c
				line = fmt.Sprintf("%s[%s%s[white]", color, color, line[1:])
			}
		}
		line = fmt.Sprintf("%s%s[white]", color, line)
		lines = append(lines, fmt.Sprintf("[grey]%d:[white]%s", lineNum, line))
	}
	log.SetText(strings.Join(lines, "\r\n"))
	return log
}

func DepTree(report format.Report) *tview.TreeView {

	var root *tview.TreeNode
	if report.DepDetailGraph != nil && report.DepDetailGraph.Name != "" {
		root = newTreeNode(report.DepDetailGraph)
	} else {
		root = tview.NewTreeNode(report.AppName).SetColor(colorPath)
	}

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

	return tree
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
		color := colorPath
		if d.Develop {
			color = colorDevDep
		}
		paths := tview.NewTreeNode("paths").SetColor(color).SetExpanded(!n.IsExpanded())
		for _, p := range d.Paths {
			paths.AddChild(tview.NewTreeNode(p).SetColor(color))
		}
		n.AddChild(paths)
	}

	// 许可证
	if len(d.Licenses) > 0 {
		color := colorLicense
		if d.Develop {
			color = colorDevDep
		}
		license := tview.NewTreeNode("license").SetColor(color).SetExpanded(!n.IsExpanded())
		for _, lic := range d.Licenses {
			license.AddChild(tview.NewTreeNode(lic.ShortName).SetColor(color))
		}
		n.AddChild(license)
	}

	// 漏洞
	if len(d.Vulnerabilities) > 0 {
		color := colorVul
		if d.Develop {
			color = colorDevDep
		}
		vulns := tview.NewTreeNode("vulns").SetColor(color).SetExpanded(!n.IsExpanded())
		for _, v := range d.Vulnerabilities {
			// 漏洞=>详细字段
			vuln := tview.NewTreeNode(v.Id).SetColor(color)
			vuln.AddChild(tview.NewTreeNode(fmt.Sprintf("name:%s", v.Name)).SetColor(color))
			vuln.AddChild(tview.NewTreeNode(fmt.Sprintf("cve:%s", v.Cve)).SetColor(color))
			vuln.AddChild(tview.NewTreeNode(fmt.Sprintf("description:%s", v.Description)).SetColor(color))
			vuln.AddChild(tview.NewTreeNode(fmt.Sprintf("suggestion:%s", v.Suggestion)).SetColor(color))
			vulns.AddChild(vuln)
		}
		n.AddChild(vulns)
	}

	return n
}
