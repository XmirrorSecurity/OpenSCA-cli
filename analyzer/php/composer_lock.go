/*
 * @Description: parse composer.lock file
 * @Date: 2021-11-26 14:50:06
 */

package php

import (
	"encoding/json"
	"sort"
	"strings"
	"util/logs"
	"util/model"
)

// composer.lock
type ComposerLock struct {
	Pkgs []struct {
		Name     string            `json:"name"`
		Version  string            `json:"version"`
		Require  map[string]string `json:"require"`
		HomePage string            `json:"homepage"`
		Source   map[string]string `json:"source"`
	} `json:"packages"`
}

// parseComposerLock parse composer.lock
func parseComposerLock(root *model.DepTree, file *model.FileInfo, direct []string) {
	lock := ComposerLock{}
	if err := json.Unmarshal(file.Data, &lock); err != nil {
		logs.Error(err)
		//return
	}
	// 记录尚无Parent的依赖
	depMap := map[string]*model.DepTree{}
	// 用来计算未被其他组件依赖的组件
	directMap := map[string]*model.DepTree{}
	for _, cps := range lock.Pkgs {
		dep := model.NewDepTree(nil)
		dep.Name = cps.Name
		dep.Version = model.NewVersion(cps.Version)
		dep.Expand = cps.Require
		dep.HomePage = cps.HomePage
		dep.DownloadLocation = strings.ReplaceAll(cps.Source["url"], ".git", "")
		depMap[cps.Name] = dep
		directMap[cps.Name] = dep
	}
	for _, dep := range depMap {
		if req, ok := dep.Expand.(map[string]string); ok {
			for n := range req {
				delete(directMap, n)
			}
		}
	}
	// 将传入的直接依赖作为根节点的子依赖
	for _, name := range direct {
		if dep, ok := depMap[name]; ok {
			dep.Parent = root
			root.Children = append(root.Children, dep)
		}
		// 避免重复添加
		delete(depMap, name)
		delete(directMap, name)
	}
	// 将未被其他组件依赖的组件作为根节点的子依赖(直接依赖)
	for _, dep := range directMap {
		root.Children = append(root.Children, dep)
	}
	sort.Slice(root.Children, func(i, j int) bool {
		return root.Children[i].Name < root.Children[j].Name
	})
	// 按层级构建依赖树
	q := []*model.DepTree{root}
	for len(q) > 0 {
		n := q[0]
		// 添加子依赖
		if req, ok := n.Expand.(map[string]string); ok {
			for name := range req {
				if dep, ok := depMap[name]; ok {
					dep.Parent = n
					n.Children = append(n.Children, dep)
					// 避免重复添加
					delete(depMap, name)
				}
			}
			if len(n.Children) > 0 {
				sort.Slice(n.Children, func(i, j int) bool {
					return n.Children[i].Name < n.Children[j].Name
				})
			}
		}
		q = append(q[1:], n.Children...)
	}
	return
}
