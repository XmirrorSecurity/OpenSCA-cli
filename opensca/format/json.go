package format

import (
	"encoding/json"

	"github.com/xmirrorsecurity/opensca-cli/opensca/model"
)

type DepJson struct {
	model.Dep
	Children []DepJson `json:"children"`
}

// Json 无重复json
func Dep2Json(dep *model.DepGraph) string {
	root := DepJson{}
	dep.Expand = root
	dep.ForEachOnce(func(n *model.DepGraph) bool {
		dj := n.Expand.(DepJson)
		dj.Dep = n.Dep
		for c := range n.Children {
			cdj := DepJson{}
			c.Expand = cdj
			dj.Children = append(dj.Children, cdj)
		}
		n.Expand = nil
		return true
	})
	data, _ := json.Marshal(root)
	return string(data)
}
