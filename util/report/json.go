package report

import (
	"encoding/json"
	"util/logs"
	"util/model"
)

// Json 获取json格式报告数据
func Json(dep *model.DepTree, taskInfo TaskInfo) []byte {
	if taskInfo.Error != nil {
		taskInfo.ErrorString = taskInfo.Error.Error()
	}
	if data, err := json.Marshal(struct {
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
