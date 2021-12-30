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

/**
 * @description: 解析package-lock.json
 * @param {*srt.DepTree} depRoot 依赖树节点
 * @param {*srt.FileData} file 文件数据
 * @return {[]*srt.DepTree} 解析出的依赖列表
 */
func parsePackageLock(depRoot *srt.DepTree, file *srt.FileData) (deps []*srt.DepTree) {
	deps = []*srt.DepTree{}
	lock := PackageLock{}
	if err := json.Unmarshal(file.Data, &lock); err != nil {
		logs.Error(err)
		return
	}
	// 记录每个依赖的信息
	depMap := map[string]*srt.DepTree{}
	// 记录每个组件的依赖
	reqMap := map[string]map[string]struct{}{}
	// 记录组件名
	nameMap := map[string]struct{}{}
	q := srt.NewQueue()
	q.Push(lock.Deps)
	for !q.Empty() {
		depmaps := q.Pop().(map[string]Dependencies)
		for name, depmap := range depmaps {
			// 统计当前组件名
			nameMap[name] = struct{}{}
			// 统计当前组件信息
			if _, ok := depMap[name]; !ok {
				dep := srt.NewDepTree(nil)
				dep.Name = name
				dep.Version = srt.NewVersion(depmap.Version)
				depMap[name] = dep
			}
			// 统计当前子依赖
			if len(depmap.Requires) > 0 && reqMap[name] == nil {
				reqMap[name] = map[string]struct{}{}
			}
			for n := range depmap.Requires {
				reqMap[name][n] = struct{}{}
			}
			// 将孙依赖添加到队列
			if len(depmap.SubDeps) > 0 {
				q.Push(depmap.SubDeps)
			}
		}
	}
	// 按组件名升序遍历每个依赖
	names := []string{}
	for name := range nameMap {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		// 按组件名升序遍历当前依赖的子依赖
		ns := []string{}
		for n := range reqMap[name] {
			ns = append(ns, n)
		}
		sort.Strings(ns)
		for _, n := range ns {
			// 将子依赖对应的依赖节点迁移到当前节点下
			if sub, ok := depMap[n]; ok && sub.Parent == nil {
				dep := depMap[name]
				sub.Parent = dep
				dep.Children = append(dep.Children, sub)
			}
		}
	}
	// 找出根节点并作为当前依赖节点的子依赖
	for _, name := range names {
		dep := depMap[name]
		if dep.Parent == nil {
			// 将当前节点信息迁移到父节点
			// depRoot.Name = dep.Name
			// depRoot.Version = dep.Version
			depRoot.Children = append(depRoot.Children, dep)
			dep.Parent = depRoot
			// for _, child := range dep.Children {
			// 	child.Parent = depRoot
			// 	depRoot.Children = append(depRoot.Children, child)
			// }
			// dep.Children = nil
			deps = append(deps, dep)
		} else {
			deps = append(deps, dep)
		}
	}
	return
}
