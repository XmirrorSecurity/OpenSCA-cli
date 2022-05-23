/*
 * @Descripation: 解析依赖
 * @Date: 2021-11-16 20:09:17
 */

package engine

import (
	"path"
	"util/filter"
	"util/model"
)

// parseDependency 解析依赖
func (e Engine) parseDependency(dirRoot *model.DirTree, depRoot *model.DepTree) *model.DepTree {
	if depRoot == nil {
		depRoot = model.NewDepTree(nil)
	}
	for _, analyzer := range e.Analyzers {
		// 遍历目录树获取要检测的文件
		files := []*model.FileInfo{}
		q := []*model.DirTree{dirRoot}
		for len(q) > 0 {
			n := q[0]
			q = q[1:]
			for _, dir := range n.DirList {
				q = append(q, n.SubDir[dir])
			}
			for _, f := range n.Files {
				if analyzer.CheckFile(f.Name) {
					files = append(files, f)
				}
			}
		}
		// 从文件中解析依赖树
		for _, d := range analyzer.ParseFiles(files) {
			depRoot.Children = append(depRoot.Children, d)
			d.Parent = depRoot
			if d.Name != "" && d.Version.Ok() {
				d.Path = path.Join(d.Path, d.Dependency.String())
			}
			// 标识为直接依赖
			d.Direct = true
			for _, c := range d.Children {
				c.Direct = true
			}
			// 填充路径
			q := []*model.DepTree{d}
			s := map[int64]struct{}{}
			for len(q) > 0 {
				n := q[0]
				n.Language = analyzer.GetLanguage()
				if _, ok := s[n.ID]; !ok {
					s[n.ID] = struct{}{}
					for _, c := range n.Children {
						if c.Path == "" {
							// 路径为空的组件在父组件路径后拼接本身依赖信息
							c.Path = path.Join(n.Path, c.Dependency.String())
						} else {
							// 路径不为空的组件在组件路径后拼接本身依赖信息
							c.Path = path.Join(c.Path, c.Dependency.String())
						}
					}
					q = append(q[1:], n.Children...)
				} else {
					q = q[1:]
				}
			}
		}
	}
	// 删除依赖树空节点
	q := []*model.DepTree{depRoot}
	for len(q) > 0 {
		n := q[0]
		q = append(q[1:], n.Children...)
		if n.Name == "" || !n.Version.Ok() {
			n.Move(n.Parent)
		}
	}
	// 校对根节点
	if depRoot.Name == "" {
		var d *model.DepTree
		for _, c := range depRoot.Children {
			if !filter.AllPkg(c.Path) {
				if d != nil {
					d = nil
					break
				} else {
					d = c
				}
			}
		}
		if d != nil {
			depRoot.Dependency = d.Dependency
			depRoot.Path = d.Path
			d.Move(depRoot)
		}
	}
	return depRoot
}
