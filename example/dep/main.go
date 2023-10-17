package main

import (
	"github.com/xmirrorsecurity/opensca-cli/opensca/logs"
	"github.com/xmirrorsecurity/opensca-cli/opensca/model"
)

func main() {

	// A->C->B
	// A->B
	A := &model.DepGraph{Name: "A"}
	B := &model.DepGraph{Name: "B"}
	C := &model.DepGraph{Name: "C"}
	A.AppendChild(C)
	A.AppendChild(B)
	C.AppendChild(B)

	logs.Info("foreach node")
	A.ForEachNode(func(p, n *model.DepGraph) bool {
		if p != nil {
			logs.Infof("%s->%s", p.Name, n.Name)
		}
		return true
	})

	logs.Info("foreach path")
	A.ForEachPath(func(p, n *model.DepGraph) bool {
		if p != nil {
			logs.Infof("%s->%s", p.Name, n.Name)
		}
		return true
	})

	logs.Infof("dep tree foreach path, sorted by addition:\n%s", A.Tree(true, false))
	logs.Infof("dep tree foreach node, sorted by addition:\n%s", A.Tree(false, false))
	logs.Infof("dep tree foreach path, sorted by name:\n%s", A.Tree(true, true))
	logs.Infof("dep tree foreach node, sorted by name:\n%s", A.Tree(false, true))

	A.Build(false, model.Lan_Java)
	logs.Infof("build by bfs, foreach path:\n%s", A.Tree(true, false))
	logs.Infof("build by bfs, foreach node:\n%s", A.Tree(false, false))

	// clear path
	A.ForEachNode(func(p, n *model.DepGraph) bool { n.Language = model.Lan_None; n.Path = ""; return true })

	A.Build(true, model.Lan_Golang)
	logs.Infof("build by dfs, foreach path:\n%s", A.Tree(true, false))
	logs.Infof("build by bfs, foreach node:\n%s", A.Tree(false, false))

}
