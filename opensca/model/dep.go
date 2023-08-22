package model

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Dep 依赖信息
type Dep struct {
	// 厂商
	Vendor string `json:"vendor"`
	// 名称
	Name string `json:"name"`
	// 版本号
	Version string `json:"version"`
	// 语言
	Language string `json:"language"`
	// 检出路径
	Path string `json:"path"`
	// 许可证
	Licenses []string `json:"licenses"`
	// 仅用于开发环境
	Develop bool `json:"develop"`
}

// DepGraph 依赖关系图
type DepGraph struct {
	// 依赖信息
	Dep
	// 父节点
	Parents map[*DepGraph]bool `json:"-"`
	// 子节点
	Children map[*DepGraph]bool `json:"-"`
	// 附加信息
	Expand any `json:"-"`
}

// AppendChild 添加子依赖
func (dep *DepGraph) AppendChild(child *DepGraph) {
	dep.Children[child] = true
	child.Parents[dep] = true
}

// RemoveChild 移除子依赖
func (dep *DepGraph) RemoveChild(child *DepGraph) {
	delete(dep.Children, child)
	delete(child.Parents, dep)
}

// Json 无重复json
func (dep *DepGraph) Json() string {
	type depJson struct {
		Dep
		Children []depJson `json:"children"`
	}
	root := depJson{}
	dep.Expand = root
	dep.ForEachOnce(func(n *DepGraph) bool {
		dj := n.Expand.(depJson)
		dj.Dep = n.Dep
		for c := range n.Children {
			cdj := depJson{}
			c.Expand = cdj
			dj.Children = append(dj.Children, cdj)
		}
		n.Expand = nil
		return true
	})
	data, _ := json.Marshal(root)
	return string(data)
}

func (dep Dep) String() string {
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
		sb.WriteString(n.Dep.String())
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
