/*
 * @Descripation: 依赖相关数据结构
 * @Date: 2021-11-03 12:11:37
 */

package model

import (
	"fmt"
	"strings"
	"sync"
	"time"
	"util/enum/language"
)

// 用于id生成
var (
	latestTime int64
	count      int64
	idMutex    sync.Mutex
)

// getId 生成一个本地唯一的id
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

// Dependency 组件依赖
type Dependency struct {
	Vendor   string        `json:"vendor,omitempty" xml:"vendor,omitempty" `
	Name     string        `json:"name,omitempty" xml:"name,omitempty" `
	Version  *Version      `json:"ver,omitempty" xml:"-" `
	Language language.Type `json:"lan,omitempty" xml:"-" `

	// 仅在生成json时赋值
	VersionStr  string `json:"version,omitempty" xml:"version,omitempty" `
	LanguageStr string `json:"language,omitempty" xml:"language,omitempty" `
}

// GetVersion 获取版本号
func (d Dependency) GetVersion() string {
	if d.Version != nil {
		return d.Version.Org
	} else {
		return d.VersionStr
	}
}

// NewDependency 创建Dependency
func NewDependency() Dependency {
	dep := Dependency{
		Vendor:   "",
		Name:     "",
		Version:  NewVersion(""),
		Language: language.None,
	}
	return dep
}

// String 获取用于展示的Dependency字符串
func (dep Dependency) String() string {
	ver := dep.VersionStr
	if dep.Version != nil {
		ver = dep.Version.Org
	}
	if len(dep.Vendor) == 0 {
		return fmt.Sprintf("[%s:%s]", dep.Name, ver)
	} else {
		return fmt.Sprintf("[%s:%s:%s]", dep.Vendor, dep.Name, ver)
	}
}

// DepTree 依赖树
type DepTree struct {
	Dependency
	// 是否为直接依赖
	Direct bool `json:"direct" xml:"direct"`
	// 依赖路径
	Path  string   `json:"-" xml:"-"`
	Paths []string `json:"paths,omitempty" xml:"paths,omitempty" `
	// 唯一的组件id，用来标识不同组件
	ID int64 `json:"-" xml:"-" `
	// 父组件
	Parent                  *DepTree `json:"-" xml:"-" `
	Vulnerabilities         []*Vuln  `json:"vulnerabilities,omitempty" xml:"vulnerabilities,omitempty" `
	IndirectVulnerabilities int      `json:"indirect_vulnerabilities,omitempty" xml:"indirect_vulnerabilities,omitempty" `
	// 许可证列表
	licenseMap map[string]struct{} `json:"-" xml:"-" `
	Licenses   []LicenseInfo       `json:"licenses,omitempty" xml:"licenses,omitempty" `
	// spdx相关字段
	CopyrightText    string `json:"copyrightText,omitempty" xml:"copyrightText,omitempty" `
	HomePage         string `json:"-" xml:"-" `
	DownloadLocation string `json:"-" xml:"-" `
	CheckSum         string `json:"-" xml:"-" `
	// 子组件
	Children []*DepTree  `json:"children,omitempty" xml:"children,omitempty" `
	Expand   interface{} `json:"-" xml:"-" `
}

// NewDepTree 创建DepTree
func NewDepTree(parent *DepTree) *DepTree {
	dep := &DepTree{
		ID:              getId(),
		Dependency:      NewDependency(),
		Vulnerabilities: []*Vuln{},
		Path:            "",
		Paths:           nil,
		Parent:          parent,
		Children:        []*DepTree{},
		licenseMap:      map[string]struct{}{},
		Licenses:        []LicenseInfo{},
		CopyrightText:   "",
	}
	if parent != nil {
		parent.Children = append(parent.Children, dep)
	}
	return dep
}

// AddLicense 添加许可证
func (dep *DepTree) AddLicense(lic LicenseInfo) {
	key := strings.TrimSpace(strings.ToLower(lic.ShortName))
	if _, ok := dep.licenseMap[key]; !ok {
		dep.licenseMap[key] = struct{}{}
		dep.Licenses = append(dep.Licenses, lic)
	}
}

// Move 将当前节点迁移到另一个节点
func (dep *DepTree) Move(other *DepTree) {
	if other == nil {
		return
	}
	if other.CopyrightText == "" {
		other.CopyrightText = dep.CopyrightText
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
	dep.Children = nil
}

// String 依赖树结构
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

		lan := dep.LanguageStr
		if lan == "" {
			lan = dep.Language.String()
		}

		lics := make([]string, len(dep.Licenses))
		for i, lic := range dep.Licenses {
			lics[i] = lic.ShortName
		}

		res += fmt.Sprintf("%s%s<%s> path:%s license:%v vulns:%v\n", strings.Repeat("\t", node.Deep), dep.Dependency, lan, dep.Path[strings.Index(dep.Path, "/")+1:], lics, vulns)
		for i := len(dep.Children) - 1; i >= 0; i-- {
			stack.Push(newNode(dep.Children[i], node.Deep+1))
		}
	}
	return res
}
