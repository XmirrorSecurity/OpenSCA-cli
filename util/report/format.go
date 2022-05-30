package report

import (
	"os"
	"strings"
	"util/enum/language"
	"util/logs"
	"util/model"
)

// 任务检查信息
type TaskInfo struct {
	AppName     string  `json:"app_name"`
	Size        int64   `json:"size"`
	StartTime   string  `json:"start_time"`
	EndTime     string  `json:"end_time"`
	CostTime    float64 `json:"cost_time"`
	Error       error   `json:"-"`
	ErrorString string  `json:"error,omitempty"`
}

// format 按照输出内容格式化(不可逆)
func format(dep *model.DepTree) {
	q := []*model.DepTree{dep}
	for len(q) > 0 {
		node := q[0]
		q = append(q[1:], node.Children...)
		if node.Language != language.None {
			node.LanguageStr = node.Language.String()
		}
		if node.Version != nil {
			node.VersionStr = node.Version.Org
		}
		node.Path = node.Path[strings.Index(node.Path, "/")+1:]
		node.Language = language.None
		node.Version = nil
	}
}

// Save 保存结果文件
func Save(data []byte, filepath string) {
	if len(data) > 0 {
		if f, err := os.Create(filepath); err != nil {
			logs.Error(err)
		} else {
			defer f.Close()
			f.Write(data)
		}
	}
}
