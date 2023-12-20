package format

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/xmirrorsecurity/opensca-cli/v3/cmd/detail"
	"github.com/xmirrorsecurity/opensca-cli/v3/opensca/logs"
)

type Report struct {
	TaskInfo TaskInfo `json:"task_info" xml:"task_info"`
	*detail.DepDetailGraph
}

type TaskInfo struct {
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
	ErrorString string `json:"error,omitempty" xml:"error,omitempty"`
}

func Save(report Report, output string) {
	for _, out := range strings.Split(output, ",") {
		logs.Infof("result save to %s", out)
		switch filepath.Ext(out) {
		case ".html":
			Html(report, out)
		case ".json":
			if strings.HasSuffix(out, ".spdx.json") {
				SpdxJson(report, out)
			} else if strings.HasSuffix(out, ".dsdx.json") {
				DsdxJson(report, out)
			} else if strings.HasSuffix(out, ".cdx.json") {
				CycloneDXJson(report, out)
			} else if strings.HasSuffix(out, ".swid.json") {
				SwidJson(report, out)
			} else {
				Json(report, out)
			}
		case ".dsdx":
			Dsdx(report, out)
		case ".spdx":
			Spdx(report, out)
		case ".xml":
			if strings.HasSuffix(out, ".spdx.xml") {
				SpdxXml(report, out)
			} else if strings.HasSuffix(out, ".dsdx.xml") {
				DsdxXml(report, out)
			} else if strings.HasSuffix(out, ".cdx.xml") {
				CycloneDXXml(report, out)
			} else if strings.HasSuffix(out, ".swid.xml") {
				SwidXml(report, out)
			} else {
				Xml(report, out)
			}
		case ".csv":
			Csv(report, out)
		case ".sqlite", ".db":
			Sqlite(report, out)
		default:
			Json(report, out)
		}
	}
}

func outWrite(out string, do func(io.Writer)) {

	if out == "" {
		do(os.Stdout)
		return
	}

	if err := os.MkdirAll(filepath.Dir(out), 0777); err != nil {
		logs.Warn(err)
		fmt.Println(err)
		return
	}

	w, err := os.Create(out)
	if err != nil {
		logs.Warn(err)
	} else {
		defer w.Close()
		do(w)
	}
}
