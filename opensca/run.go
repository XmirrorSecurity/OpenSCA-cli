package opensca

import (
	"context"
	"path"
	"time"

	"github.com/xmirrorsecurity/opensca-cli/opensca/logs"
	"github.com/xmirrorsecurity/opensca-cli/opensca/model"
	"github.com/xmirrorsecurity/opensca-cli/opensca/sca"
	"github.com/xmirrorsecurity/opensca-cli/opensca/walk"
)

func RunTask(ctx context.Context, arg *model.TaskArg) (task model.TaskResult) {

	if arg == nil {
		arg = defaultArg
	}

	if arg.Name == "" {
		arg.Name = path.Base(arg.DataOrigin)
	}

	if arg.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(arg.Timeout)*time.Second)
		if cancel != nil {
			defer cancel()
		}
	}

	task = model.TaskResult{
		AppName:   arg.Name,
		StartTime: time.Now().Format("2006-01-02 15:04:05"),
	}
	defer func(start time.Time) {
		task.CostTime = time.Since(start).Seconds()
		task.EndTime = time.Now().Format("2006-01-02 15:04:05")
	}(time.Now())

	task.Size, task.Error = walk.Walk(ctx, arg.Name, arg.DataOrigin, sca.Filter, sca.Do(ctx, func(dep *model.DepGraph) {
		logs.Info(dep)
		task.DepRoot = append(task.DepRoot, dep)
	}))

	return
}

var defaultArg = &model.TaskArg{
	DataOrigin: "./",
	Name:       "default",
}
