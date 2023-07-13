/*
 * @Description: package-lock.json 解析
 * @Date: 2021-11-25 20:25:08
 */

package javascript

import (
	"encoding/json"
	"sort"
	"util/logs"
	"util/model"
)

type dependencies struct {
	Version  string                  `json:"version"`
	Requires map[string]string       `json:"requires"`
	SubDeps  map[string]dependencies `json:"dependencies"`
}

// parseLockDepencies 解析lock依赖结构
func parseLockDepencies(root *model.DepTree, deps map[string]dependencies) {
	for name, dep := range deps {
		d := model.NewDepTree(root)
		d.Name = name
		d.Version = model.NewVersion(dep.Version)
		// 从SubDeps中解析子依赖
		parseLockDepencies(d, dep.SubDeps)
		// 从requires中删除subDep存在的组件
		for n := range dep.Requires {
			if _, exist := dep.SubDeps[n]; exist {
				delete(dep.Requires, n)
			}
		}
		d.Expand = dep.Requires
	}
}

// parsePackageLock 解析package-lock.json
func parsePackageLock(root *model.DepTree, file *model.FileInfo, direct []string) {
	lock := struct {
		Name    string                  `json:"name"`
		Version string                  `json:"version"`
		Deps    map[string]dependencies `json:"dependencies"`
	}{}
	if err := json.Unmarshal(file.Data, &lock); err != nil {
		logs.Error(err)
		//return
	}
	if lock.Name != "" {
		root.Name = lock.Name
	}
	if lock.Version != "" {
		root.Version = model.NewVersion(lock.Version)
	}
	// 解析 package-lock.json 的依赖
	parseLockDepencies(root, lock.Deps)
	// 记录未确定层级的依赖
	depMap := map[string]*model.DepTree{}
	for _, d := range root.Children {
		depMap[d.Name] = d
	}
	// direct为空则将未被其他依赖所依赖的组件设置为direct
	if len(direct) == 0 {
		directMap := map[string]struct{}{}
		for name := range depMap {
			directMap[name] = struct{}{}
		}
		q := []*model.DepTree{root}
		for len(q) > 0 {
			n := q[0]
			q = append(q[1:], n.Children...)
			if req, ok := n.Expand.(map[string]string); ok {
				for name := range req {
					delete(directMap, name)
				}
			}
		}
		for name := range directMap {
			direct = append(direct, name)
		}
	}
	// 将direct设为root的直接依赖
	root.Children = []*model.DepTree{}
	sort.Strings(direct)
	for _, name := range direct {
		if dep, exist := depMap[name]; exist {
			dep.Parent = root
			root.Children = append(root.Children, dep)
		}
		// 从depMap中删除直接依赖
		delete(depMap, name)
	}
	// 构建依赖树
	q := []*model.DepTree{root}
	for len(q) > 0 {
		n := q[0]
		// 尝试将req中的依赖与depMap记录的同名依赖匹配
		if req, ok := n.Expand.(map[string]string); ok {
			names := []string{}
			for name := range req {
				names = append(names, name)
			}
			sort.Strings(names)
			for _, name := range names {
				// depMap中匹配到依赖信息则删掉depMap中记录的依赖
				if dep, exist := depMap[name]; exist {
					n.Children = append(n.Children, dep)
					dep.Parent = n
					delete(depMap, name)
				}
			}
			n.Expand = nil
		}
		q = append(q[1:], n.Children...)
	}
	return
}
