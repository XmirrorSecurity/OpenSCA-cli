/*
 * @Descripation: 依赖相关数据结构
 * @Date: 2021-11-03 12:11:37
 */

package srt

import (
	"encoding/json"
	"fmt"
	"opensca/internal/enum/language"
	"opensca/internal/logs"
	"strings"
	"sync"
	"time"
)

// 用于id生成
var (
	latestTime int64
	count      int64
	idMutex    sync.Mutex
)

/**
 * @description: 生成一个本地唯一的id
 * @return {int64} 本地唯一的id
 */
func getId() int64 {
	idMutex.Lock()
	defer idMutex.Unlock()
	nowTime := time.Now().UnixNano() / 1e6
	if latestTime == nowTime {
		count++
	} else {
		latestTime = nowTime
		count = 0
	}
	res := nowTime
	res <<= 15
	res += count
	return res
}

/**
 * @description: 组件依赖
 */
type Dependency struct {
	Vendor   string        `json:"vendor,omitempty"`
	Name     string        `json:"name,omitempty"`
	Version  *Version      `json:"ver,omitempty"`
	Language language.Type `json:"lan,omitempty"`

	// 仅在生成json时赋值
	VersionStr  string `json:"version,omitempty"`
	LanguageStr string `json:"language,omitempty"`
}

/**
 * @description: 创建Dependency
 * @return {Dependency} 空Dependency结构
 */
func NewDependency() Dependency {
	dep := Dependency{
		Vendor:   "",
		Name:     "",
		Version:  NewVersion(""),
		Language: language.None,
	}
	return dep
}

/**
 * @description: 获取用于展示的Dependency字符串
 * @return {string} Dependency字符串
 */
func (dep Dependency) String() string {
	if len(dep.Vendor) == 0 {
		return fmt.Sprintf("[%s:%s]", dep.Name, dep.Version.Org)
	} else {
		return fmt.Sprintf("[%s:%s:%s]", dep.Vendor, dep.Name, dep.Version.Org)
	}
}

/**
 * @description: 依赖树
 */
type DepTree struct {
	Dependency
	Vulnerabilities []*Vuln `json:"vulnerabilities,omitempty"`
	// 依赖路径
	Path string `json:"path,omitempty"`
	// 唯一的组件id，用来标识不同组件
	ID int64 `json:"-"`
	// 父组件
	Parent *DepTree `json:"-"`
	// 子组件
	Children []*DepTree `json:"children,omitempty"`
	// 要排除的组件信息
	Exclusions map[string]struct{} `json:"-"`
	// 许可证列表
	licenseMap map[string]struct{} `json:"-"`
	Licenses   []string            `json:"licenses,omitempty"`
}

/**
 * @description: 创建DepTree
 * @param {*DepTree} parent 父组件，可为nil
 * @return {*DepTree} 空DepTree
 */
func NewDepTree(parent *DepTree) *DepTree {
	dep := &DepTree{
		ID:              getId(),
		Dependency:      NewDependency(),
		Vulnerabilities: []*Vuln{},
		Path:            "",
		Parent:          parent,
		Children:        []*DepTree{},
		licenseMap:      map[string]struct{}{},
		Licenses:        []string{},
		Exclusions:      map[string]struct{}{},
	}
	if parent != nil {
		parent.Children = append(parent.Children, dep)
	}
	return dep
}

/**
 * @description: 添加许可证
 * @param {string} licName 许可证名
 */
func (dep *DepTree) AddLicense(licName string) {
	key := strings.TrimSpace(strings.ToLower(licName))
	if _, ok := dep.licenseMap[key]; !ok {
		dep.licenseMap[key] = struct{}{}
		dep.Licenses = append(dep.Licenses, licName)
	}
}

/**
 * @description: 将当前节点迁移到另一个节点
 * @param {*DepTree} other 另一个依赖节点
 */
func (dep *DepTree) Move(other *DepTree) {
	if other == nil {
		return
	}
	// 从父节点中删除当前节点
	if dep.Parent != nil {
		for i, child := range dep.Parent.Children {
			if child.ID == dep.ID {
				dep.Parent.Children = append(dep.Parent.Children[:i], dep.Parent.Children[i+1:]...)
				break
			}
		}
	}
	dep.Parent = nil
	// 将子节点迁移到目标节点下
	for _, child := range dep.Children {
		child.Parent = other
		other.Children = append(other.Children, child)
	}
	// 合并Exclusion组件信息
	for exclusion := range dep.Exclusions {
		other.Exclusions[exclusion] = struct{}{}
	}
	dep.Children = nil
}

/**
 * @description: 排除Exclusion组件
 */
func (root *DepTree) Exclusion() {
	type node struct {
		Dep *DepTree
		Exc map[language.Type]map[string]struct{}
	}
	new := func(dep *DepTree, exc map[language.Type]map[string]struct{}) node {
		return node{
			Dep: dep,
			Exc: exc,
		}
	}
	q := NewQueue()
	q.Push(new(root, map[language.Type]map[string]struct{}{}))
	for !q.Empty() {
		node := q.Pop().(node)
		dep := node.Dep
		exc := node.Exc
		// 合并当前层的exclusion
		if _, ok := exc[dep.Language]; !ok {
			exc[dep.Language] = map[string]struct{}{}
		}
		for key := range dep.Exclusions {
			exc[dep.Language][key] = struct{}{}
		}
		now_exc := exc[dep.Language]
		for i := 0; i < len(dep.Children); {
			child := dep.Children[i]
			// 排除在exclusion中的组件
			key := strings.ToLower(fmt.Sprintf("%s+%s", child.Vendor, child.Name))
			if _, ok := now_exc[key]; ok && child.Language == dep.Language {
				dep.Children = append(dep.Children[:i], dep.Children[i+1:]...)
				child = nil
			} else {
				q.Push(new(child, exc))
				i++
			}
		}
	}
}

/**
 * @description: 依赖树结构
 * @return {string} 依赖树字符串
 */
func (root *DepTree) String() string {
	type node struct {
		Dep  *DepTree
		Deep int
	}
	newNode := func(dep *DepTree, deep int) *node {
		return &node{
			Dep:  dep,
			Deep: deep,
		}
	}
	// 从根节点先序遍历依赖树拼接依赖信息
	res := ""
	stack := NewStack()
	stack.Push(newNode(root, 0))
	for !stack.Empty() {
		node := stack.Pop().(*node)
		dep := node.Dep
		vulns := []string{}
		for _, v := range dep.Vulnerabilities {
			vulns = append(vulns, v.Id)
		}
		vuln := ""
		if len(vulns) > 0 {
			vuln = fmt.Sprintf(" %v", vulns)
		}
		res += fmt.Sprintf("%s%s<%s>%s%s\n", strings.Repeat("\t", node.Deep), dep.Dependency, dep.Language, dep.Path[strings.Index(dep.Path, "/")+1:], vuln)
		for i := len(dep.Children) - 1; i >= 0; i-- {
			stack.Push(newNode(dep.Children[i], node.Deep+1))
		}
	}
	return res
}

/**
 * @description: 获取用于展示结果的json数据
 * @param {string} err 错误信息
 * @return {[]byte} json数据
 */
func (dep *DepTree) Json(err string) []byte {
	// 补全依赖json信息
	q := NewQueue()
	q.Push(dep)
	for !q.Empty() {
		node := q.Pop().(*DepTree)
		for _, child := range node.Children {
			q.Push(child)
		}
		if node.Language != language.None {
			node.LanguageStr = node.Language.String()
		}
		node.VersionStr = node.Version.Org
		// 删除数据，不生成json
		node.Language = language.None
		node.Version = nil
	}
	// 生成json
	// json序列化
	if data, err := json.Marshal(struct {
		*DepTree
		Error string `json:"error,omitempty"`
	}{
		DepTree: dep,
		Error:   err,
	}); err != nil {
		logs.Error(err)
	} else {
		return data
	}
	return []byte{}
}
