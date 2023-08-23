package model

import (
	"fmt"
	"strings"
)

// DepGraph 依赖关系图
type DepGraph struct {
	// 厂商
	Vendor string
	// 名称
	Name string
	// 版本号
	Version string
	// 语言
	Language Language
	// 检出路径
	Path string
	// 许可证
	Licenses []string
	// 仅用于开发环境
	Develop bool
	// 父节点
	Parents map[*DepGraph]bool
	// 子节点
	Children map[*DepGraph]bool
	// 附加信息
	Expand any
}

// AppendChild 添加子依赖
func (dep *DepGraph) AppendChild(child *DepGraph) {
	if dep.Children == nil {
		dep.Children = map[*DepGraph]bool{}
	}
	if child.Parents == nil {
		child.Parents = map[*DepGraph]bool{}
	}
	dep.Children[child] = true
	child.Parents[dep] = true
}

// RemoveChild 移除子依赖
func (dep *DepGraph) RemoveChild(child *DepGraph) {
	delete(dep.Children, child)
	delete(child.Parents, dep)
}

func (dep *DepGraph) String() string {
	return fmt.Sprintf("[%s:%s:%s]<%s>(%s)", dep.Vendor, dep.Name, dep.Version, dep.Language, dep.Path)
}

// Tree 无重复依赖树
func (dep *DepGraph) Tree() string {

	sb := strings.Builder{}
	depSet := map[*DepGraph]bool{}
	dep.Expand = 0

	q := []*DepGraph{dep}
	for len(q) > 0 {
		l := len(q)
		n := q[l-1]
		q = q[:l-1]

		if depSet[n] {
			continue
		}
		depSet[n] = true

		deep := n.Expand.(int)
		for c := range n.Children {
			c.Expand = deep + 1
			q = append(q, c)
		}
		n.Expand = nil

		sb.WriteString(strings.Repeat("  ", deep))
		sb.WriteString(n.String())
		sb.WriteString("\n")
	}
	return sb.String()
}

// ForEach 遍历依赖图
// do: 对当前节点的操作 返回true代表继续迭代子节点
func (dep *DepGraph) ForEach(do func(n *DepGraph) bool) {
	q := []*DepGraph{dep}
	for len(q) > 0 {
		n := q[0]
		q = q[1:]
		if do(n) {
			for c := range n.Children {
				q = append(q, c)
			}
		}
	}
}

// ForEachOnce 无重复遍历依赖图
func (dep *DepGraph) ForEachOnce(do func(n *DepGraph) bool) {
	depSet := map[*DepGraph]bool{}
	dep.ForEach(func(n *DepGraph) bool {
		if depSet[n] {
			return false
		}
		depSet[n] = true
		return do(n)
	})
}
