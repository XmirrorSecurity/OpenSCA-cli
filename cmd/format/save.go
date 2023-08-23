package format

import "github.com/xmirrorsecurity/opensca-cli/cmd/detail"

type Report struct {
	ToolVersion string `json:"tool_version" xml:"tool_version" `
	// 检测目标名
	AppName string `json:"app_name" xml:"app_name" `
	// 检测文件大小
	Size int64 `json:"size" xml:"size" `
	// 任务开始时间
	StartTime string `json:"start_time" xml:"start_time" `
	// 任务结束时间
	EndTime string `json:"end_time" xml:"end_time" `
	// 任务检测耗时 单位s
	CostTime float64 `json:"cost_time" xml:"cost_time" `
	// 错误信息
	ErrorString string `json:"error" xml:"error"`
	*detail.DepDetailGraph
}

func Save(report Report, out string) {
}

// // output 输出结果
// func output(depRoot *model.DepTree, taskInfo report.TaskInfo) {
// 	taskInfo.ToolVersion = version
// 	report.Format(depRoot)
// 	// 记录依赖
// 	logs.Debug("\n" + depRoot.String())
// 	// 输出结果
// 	var reportFunc func(*model.DepTree, report.TaskInfo)
// 	out := args.Config.Out
// 	ext := path.Ext(out)
// 	switch ext {
// 	case ".html":
// 		reportFunc = report.Html
// 	case ".json":
// 		if strings.HasSuffix(out, ".spdx.json") {
// 			reportFunc = report.SpdxJson
// 		} else if strings.HasSuffix(out, ".cdx.json") {
// 			reportFunc = report.CycloneDXJson
// 		} else if strings.HasSuffix(out, ".swid.json") {
// 			out += ".zip"
// 			reportFunc = report.SwidJson
// 		} else {
// 			reportFunc = report.Json
// 		}
// 	case ".spdx":
// 		reportFunc = report.Spdx
// 	case ".xml":
// 		if strings.HasSuffix(out, ".spdx.xml") {
// 			reportFunc = report.SpdxXml
// 		} else if strings.HasSuffix(out, ".cdx.xml") {
// 			reportFunc = report.CycloneDXXml
// 		} else if strings.HasSuffix(out, ".swid.xml") {
// 			out += ".zip"
// 			reportFunc = report.SwidXml
// 		} else {
// 			reportFunc = report.Xml
// 		}
// 	case ".csv":
// 		reportFunc = report.Csv
// 	case ".sqlite", ".db":
// 		reportFunc = report.Sqlite
// 	default:
// 		reportFunc = report.Json
// 	}

// 	fmt.Println(report.Statis(depRoot, taskInfo))
// 	reportFunc(depRoot, taskInfo)

// }
