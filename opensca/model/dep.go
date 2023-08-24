package model

import (
	"fmt"
	"sort"
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
	if dep == nil || child == nil {
		return
	}
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

func (dep *DepGraph) Index() string {
	if dep.Vendor == "" {
		return fmt.Sprintf("[%s:%s]", dep.Name, dep.Version)
	}
	return fmt.Sprintf("[%s:%s:%s]", dep.Vendor, dep.Name, dep.Version)
}

func (dep *DepGraph) String() string {
	dev := ""
	if dep.Develop {
		dev = "<dev>"
	}
	return fmt.Sprintf("%s%s<%s>(%s)", dev, dep.Index(), dep.Language, dep.Path)
}

func (dep *DepGraph) IsDevelop() bool {
	if len(dep.Parents) == 0 || dep.Develop {
		return dep.Develop
	}
	for p := range dep.Parents {
		if !p.Develop {
			return false
		}
	}
	return true
}

// Tree 依赖树
// path: true=>记录全部路径 false=>记录全部节点
func (dep *DepGraph) Tree(path bool) string {

	if dep == nil {
		return ""
	}

	sb := strings.Builder{}

	dep.ForEach(true, path, func(p, n *DepGraph) bool {

		if p == nil {
			n.Expand = 0
		} else {
			n.Expand = p.Expand.(int) + 1
		}

		sb.WriteString(strings.Repeat("  ", n.Expand.(int)))
		sb.WriteString(n.String())
		sb.WriteString("\n")

		return true
	})

	return sb.String()
}

// ForEach 遍历依赖图
// deep: true=>深度优先 false=>广度优先
// path: true=>遍历所有路径 false=>遍历所有节点
// do: 对当前节点的操作 返回true代表继续迭代子节点
// do.p: 路径父节点
// do.n: 路径子节点
func (dep *DepGraph) ForEach(deep, path bool, do func(p, n *DepGraph) bool) {

	if dep == nil {
		return
	}

	var set func(p, n *DepGraph) bool
	if path {
		pathSet := map[*DepGraph]map[*DepGraph]bool{}
		set = func(p, n *DepGraph) bool {
			if _, ok := pathSet[p]; !ok {
				pathSet[p] = map[*DepGraph]bool{}
			}
			if pathSet[p][n] {
				return true
			}
			pathSet[p][n] = true
			return false
		}
	} else {
		nodeSet := map[*DepGraph]bool{}
		set = func(p, n *DepGraph) bool {
			if nodeSet[n] {
				return true
			}
			nodeSet[n] = true
			return false
		}
	}

	type pn struct {
		p *DepGraph
		n *DepGraph
	}

	q := []*pn{{nil, dep}}
	for len(q) > 0 {

		var n *pn
		if deep {
			n = q[len(q)-1]
			q = q[:len(q)-1]
		} else {
			n = q[0]
			q = q[1:]
		}

		if !do(n.p, n.n) {
			continue
		}

		var next []*DepGraph
		for c := range n.n.Children {
			next = append(next, c)
		}
		sort.Slice(next, func(i, j int) bool { return next[i].Name < next[j].Name })

		if deep {
			for i, j := 0, len(next)-1; i < j; i, j = i+1, j-1 {
				next[i], next[j] = next[j], next[i]
			}
		}

		for _, c := range next {
			if set(n.n, c) {
				continue
			}
			q = append(q, &pn{n.n, c})
		}

	}
}

// ForEachPath 遍历依赖图路径
func (dep *DepGraph) ForEachPath(do func(p, n *DepGraph) bool) {
	dep.ForEach(false, true, do)
}

// ForEachNode 遍历依赖图节点
func (dep *DepGraph) ForEachNode(do func(p, n *DepGraph) bool) {
	dep.ForEach(false, false, do)
}
