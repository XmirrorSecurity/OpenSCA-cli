/*
 * @Descripation: 依赖相关数据结构
 * @Date: 2022-11-16 10:41:37
 */

package model

import (
	"fmt"
	"util/enum/language"
)

// 参考dependency做了细微改动

// component 组件依赖
type component struct {
	Vendor   string        `json:"vendor,omitempty"`
	Name     string        `json:"name,omitempty"`
	Version  *Version      `json:"ver,omitempty"`
	Language language.Type `json:"lan,omitempty"`

	// 仅在v2请求时赋值
	LanguageStr string `json:"language"`
	//ComponentId      int64  `json:"componentId"`
	ComponentAuthor  string `json:"componentAuthor"`
	ComponentName    string `json:"componentName"`
	ComponentVersion string `json:"componentVersion"`
	FilePath         string `json:"filePath"`
}

// NewComponent 创建Component
func NewComponent() component {
	dep := component{
		Vendor:   "",
		Name:     "",
		Version:  NewVersion(""),
		Language: language.None,
	}
	return dep
}

// String 获取用于展示的Component字符串
func (dep component) String() string {
	if len(dep.Vendor) == 0 {
		return fmt.Sprintf("[%s:%s]", dep.Name, dep.Version.Org)
	} else {
		return fmt.Sprintf("[%s:%s:%s]", dep.Vendor, dep.Name, dep.Version.Org)
	}
}

// CompTree 依赖树
type CompTree struct {
	component
	// 是否为直接依赖
	Direct bool `json:"direct"`
	// 依赖路径
	Path  string   `json:"path,omitempty"`
	Paths []string `json:"paths,omitempty"`
	// 唯一的组件id，用来标识不同组件
	ID int64 `json:"id,omitempty"`
	// 父组件
	Parent                  *CompTree `json:"-"`
	Vulnerabilities         []*Vuln   `json:"vulnerabilities,omitempty"`
	IndirectVulnerabilities int       `json:"indirect_vulnerabilities,omitempty"`
	// 许可证列表
	Licenses []string `json:"licenses,omitempty"`
	// spdx相关字段
	CopyrightText string `json:"copyrightText,omitempty"`
	// 子组件
	Children []*CompTree `json:"children"`
}

// NewCompTree 创建CompTree
func NewCompTree(parent *CompTree) *CompTree {
	dep := &CompTree{
		ID:        getId(),
		component: NewComponent(),
		Path:      "",
		Paths:     nil,
		Parent:    parent,
		Children:  []*CompTree{},
	}
	if parent != nil {
		parent.Children = append(parent.Children, dep)
	}
	return dep
}

// Delete [移位法]删除子节点中指定索引的节点
func (comp *CompTree) Delete(diMap map[int]int) {
	j := 0
	for i, v := range comp.Children {
		if _, ok := diMap[i]; !ok {
			comp.Children[j] = v
			j++
		}
	}
	comp.Children = comp.Children[:j]
}
