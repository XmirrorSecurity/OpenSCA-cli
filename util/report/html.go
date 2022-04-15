package report

import (
	"errors"
	"util/logs"
	"util/model"
)

// SaveHtml 将结果保存为html文件
func SaveHtml(dep *model.DepTree, err, filepath string) {
	format(dep)
	logs.Error(errors.New("not implement"))
}
