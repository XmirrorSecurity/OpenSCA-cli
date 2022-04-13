package erlang

import (
	"opensca/internal/srt"
	"regexp"
)

// parseRebarLock parse rebar.lock file
func parseRebarLock(dep *srt.DepTree, file *srt.FileData) []*srt.DepTree {
	deps := []*srt.DepTree{}
	// pkg\s*,\s*<<"(\S+)">>\s*,\s*<<"(\S+)">>
	reg := regexp.MustCompile(`pkg\s*,\s*<<"(\S+)">>\s*,\s*<<"(\S+)">>`)
	for _, match := range reg.FindAllStringSubmatch(string(file.Data), -1) {
		sub := srt.NewDepTree(dep)
		sub.Name = match[1]
		sub.Version = srt.NewVersion(match[2])
		deps = append(deps, sub)
	}
	return deps
}
