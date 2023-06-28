package report

import (
	"encoding/xml"
	"util/logs"
	"util/model"
)

// Xml 获取xml格式报告数据
func Xml(dep *model.DepTree, taskInfo TaskInfo) []byte {
	if taskInfo.Error != nil {
		taskInfo.ErrorString = taskInfo.Error.Error()
	}
	type report struct {
		TaskInfo TaskInfo `xml:"task_info"`
		*model.DepTree
	}
	if data, err := xml.Marshal(report{
		TaskInfo: taskInfo,
		DepTree:  dep,
	}); err != nil {
		logs.Error(err)
	} else {
		return data
	}
	return []byte{}
}
