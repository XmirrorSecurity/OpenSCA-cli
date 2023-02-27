/*
 * @Descripation: 引擎入口
 * @Date: 2021-11-03 11:12:19
 */
package main

import (
	"analyzer/engine"
	"flag"
	"fmt"
	"path"
	"strings"
	"util/args"
	"util/logs"
	"util/model"
	"util/report"
)

var version string

func main() {
	args.Parse()
	if len(args.Config.Path) > 0 {
		output(engine.NewEngine().ParseFile(args.Config.Path))
	} else {
		flag.PrintDefaults()
	}
}

// output 输出结果
func output(depRoot *model.DepTree, taskInfo report.TaskInfo) {
	taskInfo.ToolVersion = version
	// 记录依赖
	logs.Debug("\n" + depRoot.String())
	// 输出结果
	var reportFunc func(*model.DepTree, report.TaskInfo) []byte
	out := args.Config.Out
	switch path.Ext(out) {
	case ".html":
		reportFunc = report.Html
	case ".json":
		if strings.HasSuffix(out, ".spdx.json") {
			reportFunc = report.SpdxJson
			break
		}
		reportFunc = report.Json
	case ".spdx":
		reportFunc = report.Spdx
	case ".xml":
		if strings.HasSuffix(out, ".spdx.xml") {
			reportFunc = report.SpdxXml
			break
		}
	default:
		reportFunc = report.Json
	}
	if args.Config.Out != "" {
		report.Save(reportFunc(depRoot, taskInfo), args.Config.Out)
	} else {
		fmt.Println(string(reportFunc(depRoot, taskInfo)))
	}
	if taskInfo.Error == nil && taskInfo.ErrorString == "" {
		coms := map[int]int{
			0: 0, 1: 0, 2: 0, 3: 0, 4: 0, 5: 0,
		}
		vuls := map[int]int{
			0: 0, 1: 0, 2: 0, 3: 0, 4: 0,
		}
		vset := map[string]bool{}
		q := []*model.DepTree{depRoot}
		for len(q) > 0 {
			n := q[0]
			risk := 5
			for _, v := range n.Vulnerabilities {
				if !vset[v.Id] {
					vset[v.Id] = true
					vuls[v.SecurityLevelId]++
					vuls[0]++
				}
				if v.SecurityLevelId < risk {
					risk = v.SecurityLevelId
				}
			}
			coms[risk]++
			coms[0]++
			q = append(q[1:], n.Children...)
		}
		fmt.Printf("\nComplete!"+
			"\nComponents:%d C:%d H:%d M:%d L:%d"+
			"\nVulnerabilities:%d C:%d H:%d M:%d L:%d\n",
			coms[0], coms[1], coms[2], coms[3], coms[4],
			vuls[0], vuls[1], vuls[2], vuls[3], vuls[4],
		)
	}
}
