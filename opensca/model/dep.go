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
	Licenses   []string
	licenseMap map[string]bool
	// 仅用于开发环境
	Develop bool
	// 直接依赖
	Direct bool
	// 父节点
	Parents []*DepGraph
	pset    map[*DepGraph]bool
	// 子节点
	Children []*DepGraph
	cset     map[*DepGraph]bool
	// 附加信息
	Expand any
}

// AppendChild 添加子依赖
func (dep *DepGraph) AppendChild(child *DepGraph) {
	if dep == nil || child == nil {
		return
	}
	if dep.cset == nil {
		dep.cset = map[*DepGraph]bool{}
	}
	if child.pset == nil {
		child.pset = map[*DepGraph]bool{}
	}
	if !dep.cset[child] {
		dep.Children = append(dep.Children, child)
		dep.cset[child] = true
	}
	if !child.pset[dep] {
		child.Parents = append(child.Parents, dep)
		child.pset[dep] = true
	}
}

// RemoveChild 移除子依赖
func (dep *DepGraph) RemoveChild(child *DepGraph) {
	for i, c := range dep.Children {
		if c == child {
			dep.Children = append(dep.Children[:i], dep.Children[i+1:]...)
			break
		}
	}
	for i, p := range child.Parents {
		if p == dep {
			child.Parents = append(child.Parents[:i], child.Parents[i+1:]...)
			break
		}
	}
	delete(dep.cset, child)
	delete(child.pset, dep)
}

func (dep *DepGraph) AppendLicense(lic string) {
	if dep.licenseMap == nil {
		dep.licenseMap = map[string]bool{}
	}
	if lic == "" {
		return
	}
	if dep.licenseMap[lic] {
		return
	}
	dep.licenseMap[lic] = true
	dep.Licenses = append(dep.Licenses, lic)
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

// Flush 更新依赖图依赖关系
func (dep *DepGraph) Flush() {

	// 删除不在dep依赖图的节点
	parentMap := map[*DepGraph]bool{}
	trueParentMap := map[*DepGraph]bool{}
	dep.ForEachPath(func(p, n *DepGraph) bool {
		// 记录所有节点的父节点
		for _, p := range n.Parents {
			parentMap[p] = true
		}
		// 记录在依赖图可遍历路径中的父节点
		if p != nil {
			trueParentMap[p] = true
		}
		return true
	})
	for p := range parentMap {
		if trueParentMap[p] {
			continue
		}
		// 删除遍历不到的父节点
		for _, c := range p.Children {
			p.RemoveChild(c)
		}
	}

	// 锁定起始组件dev
	dep.ForEachNode(func(p, n *DepGraph) bool { n.Expand = nil; return true })
	dep.ForEachNode(func(p, n *DepGraph) bool {
		// 起始开发组件状态锁定为开发组件
		if len(n.Parents) == 0 || n.Develop {
			n.Expand = struct{}{}
		}
		return true
	})

	// 传递develop
	dep.ForEachPath(func(p, n *DepGraph) bool {
		// 组件dev已锁定则跳过
		if n.Expand != nil {
			return true
		}
		// 传递父组件dev
		n.Develop = p.Develop
		// 存在任一父组件为实际引入则锁定组件dev
		if !p.Develop {
			n.Expand = struct{}{}
		}
		return true
	})

	// 去除非实际引用的关系
	dep.ForEachNode(func(p, n *DepGraph) bool {
		// 非开发组件的父组件为开发组件时 删除和开发父组件依赖关系
		if !n.Develop {
			for _, p := range n.Parents {
				if p.Develop {
					p.RemoveChild(n)
				}
			}
		}
		return true
	})
}

// Build 构建依赖图路径
// deep: 依赖路径构建顺序 true=>深度优先 false=>广度优先
// 广度优先的路径更短 如果不清楚该用什么推荐false
// lan: 更新依赖语言
func (dep *DepGraph) Build(deep bool, lan Language) {
	dep.Flush()
	dep.ForEach(deep, false, false, func(p, n *DepGraph) bool {
		// 补全路径
		if p != nil && n.Path == "" {
			n.Path = p.Path
		}
		if n.Name != "" {
			n.Path += n.Index()
		}
		// 补全语言
		if n.Language == Lan_None {
			n.Language = lan
		}
		// 直接依赖
		if len(n.Parents) == 0 || len(p.Parents) == 0 {
			n.Direct = true
		}
		return true
	})
}

// RemoveDevelop 移除develop组件
func (dep *DepGraph) RemoveDevelop() {
	dep.ForEachNode(func(p, n *DepGraph) bool {
		if n.Develop {
			for _, c := range n.Children {
				n.RemoveChild(c)
			}
			for _, p := range n.Parents {
				p.RemoveChild(n)
			}
			n = nil
			return false
		}
		return true
	})
}

// Tree 依赖树
// path: true=>记录全部路径 false=>记录全部节点
// name: true=>名称升序排序 false=>添加顺序排列
func (dep *DepGraph) Tree(path, name bool) string {

	if dep == nil {
		return ""
	}

	sb := strings.Builder{}

	dep.ForEach(true, path, name, func(p, n *DepGraph) bool {

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
// name: true=>按名称顺序迭代子节点 false=>按添加顺序迭代子节点
// do: 对当前节点的操作 返回true代表继续迭代子节点
// do.p: 路径父节点
// do.n: 路径子节点
func (dep *DepGraph) ForEach(deep, path, name bool, do func(p, n *DepGraph) bool) {

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

		if set(n.p, n.n) || !do(n.p, n.n) {
			continue
		}

		next := make([]*DepGraph, len(n.n.Children))
		copy(next, n.n.Children)

		if name {
			sort.Slice(next, func(i, j int) bool { return next[i].Name < next[j].Name })
		}

		if deep {
			for i, j := 0, len(next)-1; i < j; i, j = i+1, j-1 {
				next[i], next[j] = next[j], next[i]
			}
		}

		for _, c := range next {
			q = append(q, &pn{n.n, c})
		}

	}
}

// ForEachPath 遍历依赖图路径
func (dep *DepGraph) ForEachPath(do func(p, n *DepGraph) bool) {
	dep.ForEach(false, true, false, do)
}

// ForEachNode 遍历依赖图节点
func (dep *DepGraph) ForEachNode(do func(p, n *DepGraph) bool) {
	dep.ForEach(false, false, false, do)
}

type DepGraphMap struct {
	m     map[string]*DepGraph
	key   func(...string) string
	store func(...string) *DepGraph
}

func NewDepGraphMap(key func(...string) string, store func(...string) *DepGraph) *DepGraphMap {
	if key == nil {
		key = func(s ...string) string { return strings.Join(s, ":") }
	}
	return &DepGraphMap{key: key, store: store, m: map[string]*DepGraph{}}
}

func (s *DepGraphMap) Range(do func(k string, v *DepGraph) bool) {
	for k, v := range s.m {
		if !do(k, v) {
			break
		}
	}
}

func (s *DepGraphMap) LoadOrStore(words ...string) *DepGraph {

	if s == nil || s.key == nil || s.store == nil {
		return nil
	}

	key := s.key(words...)
	dep, ok := s.m[key]
	if !ok {
		dep = s.store(words...)
		s.m[key] = dep
	}
	return dep
}
