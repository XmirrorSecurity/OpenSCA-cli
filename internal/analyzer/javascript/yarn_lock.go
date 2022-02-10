/*
 * @Description: parse yarn.lock
 * @Date: 2022-01-20 14:28:18
 */

package javascript

import (
	"opensca/internal/srt"
	"regexp"
	"sort"
	"strings"
)

/**
 * @description: parse yarn.lock file
 * @param {*srt.DepTree} root dependency
 * @param {*srt.FileData} file yarn.lock file data
 * @return {[]*srt.DepTree} dependencies list
 */
func parseYarnLock(root *srt.DepTree, file *srt.FileData) (deps []*srt.DepTree) {
	deps = []*srt.DepTree{}
	// map[name]*DepTree
	depMap := map[string]*srt.DepTree{}
	// map[name][indirect dependencies name list]
	subMap := map[string][]string{}
	// save direct dependencies name
	directSet := map[string]struct{}{}
	depRe := regexp.MustCompile(`"([^"\s]+)@[^"\s]+"`)
	verRe := regexp.MustCompile(`"version" "([^"\s]+)"`)
	subRe := regexp.MustCompile(`"([^"\s]+)" "[^"\s]+"`)
	// traverse block
	for _, block := range strings.Split(string(file.Data), "\n\n") {
		lines := strings.Split(block, "\n")
		if len(lines) < 2 {
			continue
		}
		// match direct dependency information
		match := depRe.FindStringSubmatch(lines[0])
		name := ""
		version := ""
		if len(match) == 2 {
			name = match[1]
		} else {
			// continue without name match
			continue
		}
		// version
		match = verRe.FindStringSubmatch(block)
		if len(match) == 2 {
			version = match[1]
		}
		dep := srt.NewDepTree(nil)
		dep.Name = name
		dep.Version = srt.NewVersion(version)
		// indrect dependencies name list
		sub := []string{}
		for i, line := range lines {
			if strings.EqualFold(strings.TrimSpace(line), `dependencies:`) {
				for _, l := range lines[i+1:] {
					match = subRe.FindStringSubmatch(l)
					if len(match) == 2 {
						sub = append(sub, match[1])
					}
				}
				break
			}
		}
		depMap[dep.Name] = dep
		subMap[dep.Name] = sub
		directSet[dep.Name] = struct{}{}
	}
	// find direct dependencies
	for _, subs := range subMap {
		for _, sub := range subs {
			delete(directSet, sub)
		}
	}
	names := []string{}
	for n := range directSet {
		names = append(names, n)
	}
	sort.Strings(names)
	q := srt.NewQueue()
	// add direct dependencies
	for _, name := range names {
		dep := depMap[name]
		dep.Parent = root
		root.Children = append(root.Children, dep)
		q.Push(dep)
		deps = append(deps, dep)
	}
	// build dependency tree
	// indirecrt dependencies
	for !q.Empty() {
		dep := q.Pop().(*srt.DepTree)
		for _, name := range subMap[dep.Name] {
			if sub, ok := depMap[name]; ok && sub.Parent == nil {
				sub.Parent = dep
				dep.Children = append(dep.Children, sub)
				q.Push(sub)
			}
		}
	}
	return
}
