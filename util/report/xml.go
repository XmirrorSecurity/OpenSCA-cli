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
	if data, err := xml.Marshal(struct {
		TaskInfo TaskInfo `json:"task_info"`
		*model.DepTree
	}{
		TaskInfo: taskInfo,
		DepTree:  dep,
	}); err != nil {
		logs.Error(err)
	} else {
		return data
	}
	return []byte{}
}
