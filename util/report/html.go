package report

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"io"

	"github.com/xmirrorsecurity/opensca-cli/opensca/logs"
	"github.com/xmirrorsecurity/opensca-cli/util/model"
)

//go:embed html_tpl
var index []byte

// Html 获取html格式报告数据
func Html(dep *model.DepTree, taskInfo TaskInfo) {
	// html组件字段
	type htmlDep struct {
		*model.DepTree
		SecId  int         `json:"security_level_id,omitempty"`
		Statis map[int]int `json:"vuln_statis"`
	}
	deps := []htmlDep{}
	// html统计信息
	type htmlStatis struct {
		Component map[int]int `json:"component"`
		Vuln      map[int]int `json:"vuln"`
	}
	statis := htmlStatis{
		Component: map[int]int{},
		Vuln:      map[int]int{},
	}
	vulnMap := map[string]int{}
	// 遍历所有组件
	q := []*model.DepTree{dep}
	for len(q) > 0 {
		n := q[0]
		q = append(q[1:], n.Children...)
		// 删除不需要的数据
		n.Children = nil
		n.IndirectVulnerabilities = 0
		// 计算组件风险
		secid := 5
		vuln_statis := map[int]int{}
		for _, v := range n.Vulnerabilities {
			vulnMap[v.Id] = v.SecurityLevelId
			vuln_statis[v.SecurityLevelId]++
			if secid > v.SecurityLevelId {
				secid = v.SecurityLevelId
			}
		}
		if n.Name != "" {
			statis.Component[secid]++
			deps = append(deps, htmlDep{
				n,
				secid,
				vuln_statis,
			})
		}
	}
	// 统计漏洞风险
	for _, secid := range vulnMap {
		statis.Vuln[secid]++
	}
	// 生成html报告需要的json数据
	if data, err := json.Marshal(struct {
		TaskInfo   TaskInfo   `json:"task_info"`
		Statis     htmlStatis `json:"statis"`
		Components []htmlDep  `json:"components"`
	}{
		TaskInfo:   taskInfo,
		Statis:     statis,
		Components: deps,
	}); err != nil {
		logs.Warn(err)
	} else {
		// 替换模板数据
		outWrite(func(w io.Writer) {
			w.Write(bytes.Replace(index, []byte(`"此处填充json数据"`), data, 1))
		})
		return
	}
}
