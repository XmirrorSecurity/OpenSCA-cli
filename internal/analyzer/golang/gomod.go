/*
 * @Description: parse go.mod/go.sum
 * @Date: 2022-02-10 16:16:24
 */

package golang

import (
	"opensca/internal/srt"
	"regexp"
	"strings"
)

/**
 * @description: parse go.mod
 * @param {*srt.DepTree} dep dependency node
 * @param {*srt.FileData} file go.mod file data
 * @return {[]*srt.DepTree} dependencies list
 */
func parseGomod(dep *srt.DepTree, file *srt.FileData) (deps []*srt.DepTree) {
	deps = []*srt.DepTree{}
	for _, match := range regexp.MustCompile(`(\S*)\s+(v[\d\w\-+.]*)[\s\n]`).FindAllStringSubmatch(string(file.Data), -1) {
		if len(match) != 3 {
			continue
		}
		sub := srt.NewDepTree(dep)
		sub.Name = strings.Trim(match[1], `'"`)
		sub.Version = srt.NewVersion(match[2])
		deps = append(deps, sub)
	}
	return deps
}

/**
 * @description: parse go.sum
 * @param {*srt.DepTree} dep dependency node
 * @param {*srt.FileData} file go.sum file data
 * @return {[]*srt.DepTree} dependencies list
 */
func parseGosum(dep *srt.DepTree, file *srt.FileData) (deps []*srt.DepTree) {
	deps = parseGomod(dep, file)
	exist := map[string]struct{}{}
	for _, dep := range deps {
		exist[dep.Name] = struct{}{}
	}
	for _, match := range regexp.MustCompile(`(\S*)\s+(v[\d\w\-+.]*)/go.mod[\s\n]`).FindAllStringSubmatch(string(file.Data), -1) {
		if len(match) != 3 {
			continue
		}
		if _, ok := exist[match[1]]; ok {
			continue
		}
		sub := srt.NewDepTree(dep)
		sub.Name = strings.Trim(match[1], `'"`)
		sub.Version = srt.NewVersion(match[2])
		exist[sub.Name] = struct{}{}
		deps = append(deps, sub)
	}
	return deps
}
