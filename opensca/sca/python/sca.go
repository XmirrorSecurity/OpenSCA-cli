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

func (sca Sca) Sca(ctx context.Context, parent *model.File, files []*model.File) []*model.DepGraph {

	path2dir := func(relpath string) string { return path.Dir(strings.ReplaceAll(relpath, `\`, `/`)) }

	lockSet := map[string]bool{}
	for _, file := range files {
		if filter.PythonPipfileLock(file.Relpath()) {
			lockSet[path2dir(file.Relpath())] = true
		}
	}

	var roots []*model.DepGraph
	for _, file := range files {
		if filter.PythonPipfile(file.Relpath()) {
			if !lockSet[path2dir(file.Relpath())] {
				roots = append(roots, ParsePipfile(file))
			}
		} else if filter.PythonPipfileLock(file.Relpath()) {
			roots = append(roots, ParsePipfileLock(file))
		} else if filter.PythonRequirementsIn(file.Relpath()) {
			roots = append(roots, ParseRequirementIn(file))
		} else if filter.PythonRequirementsTxt(file.Relpath()) {
			roots = append(roots, ParseRequirementTxt(file))
		} else if filter.PythonSetup(file.Relpath()) {
			roots = append(roots, ParseSetup(file))
		}
	}
	return roots
}
