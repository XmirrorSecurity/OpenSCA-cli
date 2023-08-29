package opensca

import (
	"context"
	"path/filepath"
	"time"

	"github.com/xmirrorsecurity/opensca-cli/opensca/model"
	"github.com/xmirrorsecurity/opensca-cli/opensca/sca"
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
	// 调用的sca(为空时使用默认配置)
	Sca []sca.Sca
	// 额外的文件过滤函数
	ExtractFileFilter walk.ExtractFileFilter
	// 额外的文件处理函数
	WalkFileFunc walk.WalkFileFunc
}

func RunTask(ctx context.Context, arg *TaskArg) (deps []*model.DepGraph, err error) {

	if arg == nil {
		arg = &TaskArg{DataOrigin: "./"}
	}

	if arg.ExtractFileFilter == nil {
		arg.ExtractFileFilter = walk.IsCompressFile
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

	err = walk.Walk(ctx, arg.Name, arg.DataOrigin, func(relpath string) bool {
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
		if arg.WalkFileFunc != nil {
			arg.WalkFileFunc(parent, files)
		}
		for _, sca := range arg.Sca {
			for _, dep := range sca.Sca(ctx, parent, files) {
				dep.Build(false, sca.Language())
				deps = append(deps, dep)
			}
		}
	})
	return
}
