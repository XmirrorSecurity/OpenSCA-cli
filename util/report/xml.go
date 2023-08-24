package report

import (
	"encoding/xml"
	"io"

	"github.com/xmirrorsecurity/opensca-cli/opensca/logs"
	"github.com/xmirrorsecurity/opensca-cli/util/model"
)

// Xml 获取xml格式报告数据
func Xml(dep *model.DepTree, taskInfo TaskInfo) {
	if taskInfo.Error != nil {
		taskInfo.ErrorString = taskInfo.Error.Error()
	}
	type report struct {
		TaskInfo TaskInfo `xml:"task_info"`
		*model.DepTree
	}
	outWrite(func(w io.Writer) {
		err := xml.NewEncoder(w).Encode(report{
			TaskInfo: taskInfo,
			DepTree:  dep,
		})
		if err != nil {
			logs.Error(err)
		}
	})
}
