/*
 * @Descripation: 解析依赖
 * @Date: 2021-11-16 20:09:17
 */

package engine

import (
	"path"
	"util/bar"
	"util/model"
)

// parseDependency 解析依赖
func (e Engine) parseDependency(dirRoot *model.DirTree, depRoot *model.DepTree) *model.DepTree {
	type node struct {
		Dir *model.DirTree
		Dep *model.DepTree
	}
	newNode := func(dirRoot *model.DirTree, depRoot *model.DepTree) *node {
		return &node{
			Dir: dirRoot,
			Dep: depRoot,
		}
	}
	if depRoot == nil {
		depRoot = model.NewDepTree(nil)
	}
	queue := model.NewQueue()
	for _, analyzer := range e.Analyzers {
		// 将根目录添加到队列
		queue.Push(newNode(dirRoot, depRoot))
		for !queue.Empty() {
			node := queue.Pop().(*node)
			// 解析文件
			for _, file := range analyzer.FilterFile(node.Dir, node.Dep) {
				q := model.NewQueue()
				// parse dependencies
				for _, dep := range analyzer.ParseFile(node.Dir, node.Dep, file) {
					bar.Dependency.Add(1)
					dep.Path = path.Join(node.Dir.Path, path.Base(file.Name), dep.Dependency.String())
					dep.Language = analyzer.GetLanguage()
					q.Push(dep)
				}
				// add indirect dependencies infomation(path, language)
				for !q.Empty() {
					now := q.Pop().(*model.DepTree)
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
				queue.Push(newNode(node.Dir.SubDir[dir], model.NewDepTree(node.Dep)))
			}
		}
	}
	e.javaAnalyzer.BuildTree(depRoot)
	// 删除依赖树空节点
	queue.Push(depRoot)
	for !queue.Empty() {
		node := queue.Pop().(*model.DepTree)
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
