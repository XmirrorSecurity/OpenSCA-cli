package opensca

import (
	"context"
	"path/filepath"
	"time"

	"github.com/xmirrorsecurity/opensca-cli/opensca/model"
	"github.com/xmirrorsecurity/opensca-cli/opensca/sca"
	"github.com/xmirrorsecurity/opensca-cli/opensca/sca/filter"
	"github.com/xmirrorsecurity/opensca-cli/opensca/walk"
)

// 任务检测参数
type TaskArg struct {

	// 检测数据源 文件路径或url 兼容http(s)|ftp|file
	DataOrigin string
	// 检测对象名称 用于结果展示 缺省时取DataOrigin尾单词
	Name string
	// 超时时间 单位s
	Timeout int
	// 使用的sca(为空时使用默认配置)
	Sca []sca.Sca

	// 额外的文件过滤函数 默认为压缩文件名过滤函数
	ExtractFileFilter walk.ExtractFileFilter
	// 额外的结果回调函数
	ResCallFunc model.ResCallback
}

type TaskResult struct {
	// 任务参数
	Arg *TaskArg
	// 检出组件
	Deps []*model.DepGraph
	// 错误信息
	Error error
	// 任务开始时间
	Start time.Time
	// 任务结束时间
	End time.Time
	// 检测文件大小
	Size int64
}

// RunTask 运行检测任务
// arg: 任务参数
func RunTask(ctx context.Context, arg *TaskArg) (result TaskResult) {

	result.Start = time.Now()
	defer func() {
		result.End = time.Now()
		result.Arg = arg
	}()

	if arg == nil {
		arg = &TaskArg{DataOrigin: "./"}
	}

	if arg.ExtractFileFilter == nil {
		arg.ExtractFileFilter = filter.CompressFile
	}

	if arg.Name == "" {
		arg.Name = filepath.Base(arg.DataOrigin)
	}

	if arg.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(arg.Timeout)*time.Second)
		if cancel != nil {
			defer cancel()
		}
	}

	if len(arg.Sca) == 0 {
		arg.Sca = sca.AllSca
	}

	result.Size, result.Error = walk.Walk(ctx, arg.Name, arg.DataOrigin, func(relpath string) bool {

		if arg.ExtractFileFilter != nil && arg.ExtractFileFilter(relpath) {
			return true
		}

		for _, sca := range arg.Sca {
			if sca.Filter(relpath) {
				return true
			}
		}

		return false

	}, func(parent *model.File, files []*model.File) {

		for _, sca := range arg.Sca {
			sca.Sca(ctx, parent, files, func(file *model.File, root ...*model.DepGraph) {
				for _, dep := range root {
					if dep == nil {
						continue
					}
					dep.Build(false, sca.Language())
					result.Deps = append(result.Deps, dep)
					if arg.ResCallFunc != nil {
						arg.ResCallFunc(file, dep)
					}
				}
			})
		}

	})

	return result
}
