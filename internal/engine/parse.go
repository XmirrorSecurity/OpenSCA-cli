/*
 * @Descripation: 解析依赖
 * @Date: 2021-11-16 20:09:17
 */

package engine

import (
	"opensca/internal/bar"
	"opensca/internal/srt"
	"path"
)

/**
 * @description: 解析依赖
 * @param {*srt.DirTree} dirRoot 目录树根节点
 * @param {*srt.DepTree} depRoot 依赖树根节点
 * @return {*srt.DepTree} 依赖树根节点
 */
func (e Engine) parseDependency(dirRoot *srt.DirTree, depRoot *srt.DepTree) *srt.DepTree {
	type node struct {
		Dir *srt.DirTree
		Dep *srt.DepTree
	}
	newNode := func(dirRoot *srt.DirTree, depRoot *srt.DepTree) *node {
		return &node{
			Dir: dirRoot,
			Dep: depRoot,
		}
	}
	if depRoot == nil {
		depRoot = srt.NewDepTree(nil)
	}
	queue := srt.NewQueue()
	for _, analyzer := range e.Analyzers {
		// 将根目录添加到队列
		queue.Push(newNode(dirRoot, depRoot))
		for !queue.Empty() {
			node := queue.Pop().(*node)
			// 解析文件
			for _, file := range analyzer.FilterFile(node.Dir, node.Dep) {
				q := srt.NewQueue()
				// parse dependencies
				for _, dep := range analyzer.ParseFile(node.Dir, node.Dep, file) {
					bar.Dependency.Add(1)
					dep.Path = path.Join(node.Dir.Path, path.Base(file.Name), dep.Dependency.String())
					dep.Language = analyzer.GetLanguage()
					q.Push(dep)
				}
				// add indirect dependencies infomation(path, language)
				for !q.Empty() {
					now := q.Pop().(*srt.DepTree)
					for _, child := range now.Children {
						bar.Dependency.Add(1)
						child.Path = path.Join(now.Path, child.Dependency.String())
						child.Language = analyzer.GetLanguage()
						q.Push(child)
					}
				}
			}
			// 将子目录添加到队列
			for _, dir := range node.Dir.DirList {
				queue.Push(newNode(node.Dir.SubDir[dir], srt.NewDepTree(node.Dep)))
			}
		}
	}
	// 删除依赖树空节点
	queue.Push(depRoot)
	for !queue.Empty() {
		node := queue.Pop().(*srt.DepTree)
		for _, child := range node.Children {
			queue.Push(child)
		}
		if len(node.Name) == 0 || !node.Version.Ok() {
			node.Move(node.Parent)
		}
	}
	// 排除exlusion组件
	depRoot.Exclusion()
	return depRoot
}
