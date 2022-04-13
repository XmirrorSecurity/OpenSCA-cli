/*
 * @Descripation: package-lock.json 解析
 * @Date: 2021-11-25 20:25:08
 */

package javascript

import (
	"encoding/json"
	"opensca/internal/logs"
	"opensca/internal/srt"
	"sort"
)

type Dependencies struct {
	Version  string                  `json:"version"`
	Requires map[string]string       `json:"requires"`
	SubDeps  map[string]Dependencies `json:"dependencies"`
}

type PackageLock struct {
	Deps map[string]Dependencies `json:"dependencies"`
}

// parsePackageLock 解析package-lock.json
func parsePackageLock(depRoot *srt.DepTree, file *srt.FileData) (deps []*srt.DepTree) {
	deps = []*srt.DepTree{}
	lock := PackageLock{}
	if err := json.Unmarshal(file.Data, &lock); err != nil {
		logs.Error(err)
		return
	}
	// 记录每个依赖的信息
	depMap := map[string]*srt.DepTree{}
	// 记录是否为直接依赖
	directSet := map[string]struct{}{}
	// 记录每个组件的依赖
	reqMap := map[string][]string{}
	// 记录组件名
	nameMap := map[string]struct{}{}
	q := srt.NewQueue()
	q.Push(lock.Deps)
	for !q.Empty() {
		depmaps := q.Pop().(map[string]Dependencies)
		for name, depmap := range depmaps {
			// 统计当前组件名
			nameMap[name] = struct{}{}
			directSet[name] = struct{}{}
			// 统计当前组件信息
			if d, ok := depMap[name]; !ok {
				dep := srt.NewDepTree(nil)
				dep.Name = name
				dep.Version = srt.NewVersion(depmap.Version)
				depMap[name] = dep
			} else {
				newver := srt.NewVersion(depmap.Version)
				if d.Version.Less(newver) {
					d.Version = newver
				} else {
					continue
				}
			}
			// 统计当前子依赖
			if len(depmap.Requires) > 0 && reqMap[name] == nil {
				reqMap[name] = []string{}
			}
			ns := []string{}
			for n := range depmap.Requires {
				ns = append(ns, n)
			}
			reqMap[name] = ns
			// 将孙依赖添加到队列
			if len(depmap.SubDeps) > 0 {
				q.Push(depmap.SubDeps)
			}
		}
	}
	// 去掉被依赖的组件
	for _, subs := range reqMap {
		for _, sub := range subs {
			delete(directSet, sub)
		}
	}
	names := []string{}
	for name := range directSet {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		dep := depMap[name]
		depRoot.Children = append(depRoot.Children, dep)
		dep.Parent = depRoot
		deps = append(deps, dep)
		q.Push(dep)
	}
	for !q.Empty() {
		dep := q.Pop().(*srt.DepTree)
		subs := reqMap[dep.Name]
		sort.Strings(subs)
		for _, name := range subs {
			if sub, ok := depMap[name]; ok && sub.Parent == nil {
				sub.Parent = dep
				dep.Children = append(dep.Children, sub)
				q.Push(sub)
			}
		}
	}
	return
}
