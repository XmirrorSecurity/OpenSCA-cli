package python

import (
	"context"
	"path"
	"strings"

	"github.com/xmirrorsecurity/opensca-cli/opensca/model"
	"github.com/xmirrorsecurity/opensca-cli/opensca/sca/filter"
)

type Sca struct{}

func (sca Sca) Language() model.Language {
	return model.Lan_Python
}

func (sca Sca) Filter(relpath string) bool {
	return filter.PythonPipfileLock(relpath) ||
		filter.PythonPipfile(relpath) ||
		filter.PythonRequirementsIn(relpath) ||
		filter.PythonRequirementsTxt(relpath) ||
		filter.PythonSetup(relpath)
}

func (sca Sca) Sca(ctx context.Context, parent *model.File, files []*model.File, call model.ResCallback) {

	path2dir := func(relpath string) string { return path.Dir(strings.ReplaceAll(relpath, `\`, `/`)) }

	// 记录存在lock文件的目录
	lockSet := map[string]bool{}
	for _, file := range files {
		if filter.PythonPipfileLock(file.Relpath()) {
			lockSet[path2dir(file.Relpath())] = true
		}
	}

	// 记录使用pipenv解析过的目录
	pipSet := map[string]bool{}
	// 尝试使用pipenv解析
	for _, file := range files {
		if pipSet[path2dir(file.Relpath())] {
			continue
		}
		if filter.PythonPipfile(file.Relpath()) ||
			filter.PythonRequirementsTxt(file.Relpath()) {
		} else {
			continue
		}
		root := ParsePythonWithEnv(ctx, file)
		if root == nil {
			continue
		}
		call(file, root)
		pipSet[path2dir(file.Relpath())] = true
	}

	// 静态解析
	for _, file := range files {
		if pipSet[path2dir(file.Relpath())] {
			continue
		}
		if filter.PythonPipfile(file.Relpath()) {
			if !lockSet[path2dir(file.Relpath())] {
				call(file, ParsePipfile(file))
			}
		} else if filter.PythonPipfileLock(file.Relpath()) {
			call(file, ParsePipfileLock(file))
		} else if filter.PythonRequirementsIn(file.Relpath()) {
			call(file, ParseRequirementIn(file))
		} else if filter.PythonRequirementsTxt(file.Relpath()) {
			call(file, ParseRequirementTxt(file))
		} else if filter.PythonSetup(file.Relpath()) {
			call(file, ParseSetup(file))
		}
	}
}
