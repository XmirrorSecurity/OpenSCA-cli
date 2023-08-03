package report

import (
	"fmt"
	"io"
	"util/model"
)

// Csv csv格式报告数据
func Csv(dep *model.DepTree, taskInfo TaskInfo) {

	if taskInfo.Error != nil {
		taskInfo.ErrorString = taskInfo.Error.Error()
	}

	table := "Name, Version, Vendor, License, Langauge, PURL\n"

	// 遍历所有组件树，提取关键字段
	q := []*model.DepTree{dep}
	for len(q) > 0 {
		n := q[0]
		q = append(q[1:], n.Children...)

		licenseTxt := ""
		if len(n.Licenses) > 0 {
			licenseTxt = n.Licenses[0].ShortName
		}

		if n.Name != "" {
			table = table + fmt.Sprintf("%s,%s,%s,%s,%s,%s\n", n.Name, n.VersionStr, n.Vendor, licenseTxt, n.LanguageStr, n.Purl())
		}

	}

	outWrite(func(w io.Writer) {
		w.Write([]byte(table))
	})

}
