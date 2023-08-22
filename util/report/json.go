package report

import (
	"encoding/json"
	"io"

	"github.com/xmirrorsecurity/opensca-cli/util/logs"
	"github.com/xmirrorsecurity/opensca-cli/util/model"
)

// Json 获取json格式报告数据
func Json(dep *model.DepTree, taskInfo TaskInfo) {
	if taskInfo.Error != nil {
		taskInfo.ErrorString = taskInfo.Error.Error()
	}
	outWrite(func(w io.Writer) {
		err := json.NewEncoder(w).Encode(struct {
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
