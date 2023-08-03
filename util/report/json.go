package report

import (
	"encoding/json"
	"io"
	"util/logs"
	"util/model"
)

// Json 获取json格式报告数据
func Json(dep *model.DepTree, taskInfo TaskInfo) {
	if taskInfo.Error != nil {
		taskInfo.ErrorString = taskInfo.Error.Error()
	}
	outWrite(func(w io.Writer) {
		jsonEncoder := json.NewEncoder(w)
		jsonEncoder.SetIndent("", "    ")
		err := jsonEncoder.Encode(struct {
			TaskInfo TaskInfo `json:"task_info"`
			*model.DepTree
		}{
			TaskInfo: taskInfo,
			DepTree:  dep,
		})
		if err != nil {
			logs.Error(err)
		}
	})
}
