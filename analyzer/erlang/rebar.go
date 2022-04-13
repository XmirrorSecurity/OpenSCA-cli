package erlang

import (
	"regexp"
	"util/model"
)

// parseRebarLock parse rebar.lock file
func parseRebarLock(dep *model.DepTree, file *model.FileData) []*model.DepTree {
	deps := []*model.DepTree{}
	// pkg\s*,\s*<<"(\S+)">>\s*,\s*<<"(\S+)">>
	reg := regexp.MustCompile(`pkg\s*,\s*<<"(\S+)">>\s*,\s*<<"(\S+)">>`)
	for _, match := range reg.FindAllStringSubmatch(string(file.Data), -1) {
		sub := model.NewDepTree(dep)
		sub.Name = match[1]
		sub.Version = model.NewVersion(match[2])
		deps = append(deps, sub)
	}
	return deps
}
