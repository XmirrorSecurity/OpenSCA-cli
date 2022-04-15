package report

import (
	"encoding/json"
	"os"
	"util/logs"
	"util/model"
)

// SaveJson 将结果保存为json文件
func SaveJson(dep *model.DepTree, err string, filepath string) {
	if data := Json(dep, err); len(data) > 0 {
		if f, err := os.Create(filepath); err != nil {
			logs.Error(err)
		} else {
			defer f.Close()
			f.Write(data)
		}
	}
}

// Json 获取用于展示结果的json数据
func Json(dep *model.DepTree, err string) []byte {
	format(dep)
	if data, err := json.Marshal(struct {
		*model.DepTree
		Error string `json:"error,omitempty"`
	}{
		DepTree: dep,
		Error:   err,
	}); err != nil {
		logs.Error(err)
	} else {
		return data
	}
	return []byte{}
}
