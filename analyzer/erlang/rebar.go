package erlang

import (
	"regexp"
	"util/model"
)

// parseRebarLock parse rebar.lock file
func parseRebarLock(dep *model.DepTree, file *model.FileData) []*model.DepTree {
	deps := []*model.DepTree{}
	// <<"([\w\d]+)">>\S*?pkg,<<"[\w\d]+">>,<<"([.\d]+)">>
	reg := regexp.MustCompile(`<<"([\w\d]+)">>\S*?pkg,<<"[\w\d]+">>,<<"([.\d]+)">>`)
	for _, match := range reg.FindAllStringSubmatch(string(file.Data), -1) {
		sub := model.NewDepTree(dep)
		sub.Name = match[1]
		sub.Version = model.NewVersion(match[2])
		deps = append(deps, sub)
	}
	return deps
}
