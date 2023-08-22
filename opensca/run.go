package opensca

import (
	"context"
	"time"

	"github.com/xmirrorsecurity/opensca-cli/opensca/logs"
	"github.com/xmirrorsecurity/opensca-cli/opensca/model"
	"github.com/xmirrorsecurity/opensca-cli/opensca/sca"
	"github.com/xmirrorsecurity/opensca-cli/opensca/walk"
)

func RunTask(ctx context.Context, arg *TaskArg) {

	if arg == nil {
		arg = defaultArg
	}

	if arg.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(arg.Timeout)*time.Second)
		if cancel != nil {
			defer cancel()
		}
	}

	walk.Walk(ctx, arg.Name, arg.DataOrigin, sca.Filter, sca.Do(ctx, func(dep *model.DepGraph) {
		logs.Info(dep)
	}))
}

// 任务检测参数
type TaskArg struct {
	// 检测数据源 文件路径或url 兼容http(s)|ftp|file
	DataOrigin string
	// 检测对象名称 用于结果展示 缺省时取DataOrigin尾单词
	Name string
	// 超时时间 单位s
	Timeout int
}

var defaultArg = &TaskArg{
	DataOrigin: "./",
}
