/*
 * @Descripation: 依赖相关数据结构
 * @Date: 2021-11-03 12:11:37
 */

package model

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"
	"util/enum/language"
	"util/filter"
	"util/logs"
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
	Vendor   string        `json:"vendor,omitempty"`
	Name     string        `json:"name,omitempty"`
	Version  *Version      `json:"ver,omitempty"`
	Language language.Type `json:"lan,omitempty"`

	// 仅在生成json时赋值
	VersionStr  string `json:"version,omitempty"`
	LanguageStr string `json:"language,omitempty"`
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
	if len(dep.Vendor) == 0 {
		return fmt.Sprintf("[%s:%s]", dep.Name, dep.Version.Org)
	} else {
		return fmt.Sprintf("[%s:%s:%s]", dep.Vendor, dep.Name, dep.Version.Org)
	}
}

// DepTree 依赖树
type DepTree struct {
	Dependency
	// 是否为直接依赖
	Direct bool `json:"direct"`
	// 依赖路径
	Path  string   `json:"path,omitempty"`
	Paths []string `json:"paths,omitempty"`
	// 唯一的组件id，用来标识不同组件
	ID int64 `json:"id,omitempty"`
	// 父组件
	Parent                  *DepTree `json:"-"`
	Vulnerabilities         []*Vuln  `json:"vulnerabilities,omitempty"`
	IndirectVulnerabilities int      `json:"indirect_vulnerabilities,omitempty"`
	// 许可证列表
	licenseMap map[string]struct{} `json:"-"`
	Licenses   []string            `json:"licenses,omitempty"`
	// spdx相关字段
	CopyrightText    string `json:"copyrightText,omitempty"`
	HomePage         string `json:"-"`
	DownloadLocation string `json:"-"`
	CheckSum         string `json:"-"`
	// 子组件
	Children []*DepTree  `json:"children,omitempty"`
	Expand   interface{} `json:"-"`
}
type CheckSum struct {
	Algorithm string `json:"algorithm,omitempty"`
	Value     string `json:"value,omitempty"`
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
		Licenses:        []string{},
		CopyrightText:   "",
	}
	if parent != nil {
		parent.Children = append(parent.Children, dep)
	}
	return dep
}

// AddLicense 添加许可证
func (dep *DepTree) AddLicense(licName string) {
	key := strings.TrimSpace(strings.ToLower(licName))
	if _, ok := dep.licenseMap[key]; !ok {
		dep.licenseMap[key] = struct{}{}
		dep.Licenses = append(dep.Licenses, licName)
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

// ToDetectComponents 转为检测依赖组件的格式
func (root *DepTree) ToDetectComponents() (comp *CompTree) {
	// 利用json.Unmarshal来实现深拷贝
	comp = NewCompTree(nil)
	if data, err := json.Marshal(root); err != nil {
		logs.Error(err)
	} else {
		err = json.Unmarshal(data, &comp)
		if err != nil {
			logs.Error(err)
		} else {
			// 参考format，改了下顺序，先去重再转换
			// 恢复转换后缺失的父节点
			for _, c := range comp.Children {
				c.Parent = comp
			}
			// 去重
			q := []*CompTree{comp}
			dm := map[string]*CompTree{}
			for len(q) > 0 {
				n := q[0]
				q = append(q[1:], n.Children...)
				// 去重
				k := fmt.Sprintf("%s:%s@%s#%s", n.Vendor, n.Name, n.Version.Org, strings.ToLower(n.Language.String()))
				if d, ok := dm[k]; !ok {
					dm[k] = n
				} else {
					// 已存在相同组件，但是某些字段可能不一样

					// 当d.Path为空或者等于n.Path时才进行合并处理
					if d.Path == "" {
						if n.Path != "" {
							d.Path = n.Path
						}
						if n.Direct {
							d.Direct = n.Direct
						}
					} else if d.Path != n.Path {
						continue
					}
					// 从父组件中移除当前组件
					if n.Parent != nil {
						for i, c := range n.Parent.Children {
							if c.ID == n.ID {
								n.Parent.Children = append(n.Parent.Children[:i], n.Parent.Children[i+1:]...)
								break
							}
						}
					}
					// 将当前组件的子组件转移到已存在组件的子依赖中
					d.Children = append(d.Children, n.Children...)
					for _, c := range n.Children {
						c.Parent = d
					}
				}
			}

			//应该只有pom的解析存在不需要一级节点这种情况。
			//先将二级子节点复制到一级，并记录需要删除的索引
			deleteIndex := make(map[int]int)
			for i, c := range comp.Children {
				if filter.JavaPom(strings.TrimSuffix(c.Path, "/"+c.String())) {
					deleteIndex[i] = i
					c.Parent.Children = append(c.Parent.Children, c.Children...)
				}
			}
			//再删除对应的一级节点
			if len(deleteIndex) > 0 {
				comp.Delete(deleteIndex)
			}

			// 保留要导出的数据
			q = []*CompTree{comp}
			for len(q) > 0 {
				n := q[0]
				q = append(q[1:], n.Children...)
				if n.Language != language.None {
					n.LanguageStr = strings.ToLower(n.Language.String())
					n.Language = language.None
				}
				if n.Version != nil {
					n.ComponentVersion = n.Version.Org
					n.Version = nil
				}
				if n.Vendor != "" {
					n.ComponentAuthor = n.Vendor
					n.Vendor = ""
				}
				if n.Name != "" {
					n.ComponentName = n.Name
					n.Name = ""
				}
				if n.Path != "" {
					n.FilePath = n.Path
					n.Path = ""
				}
				//n.ComponentId = n.ID
				// 不展示的字段置空
				//n.ID = 0
				n.Paths = nil
				n.Licenses = nil
				n.CopyrightText = ""
				n.Vulnerabilities = nil
				n.IndirectVulnerabilities = 0
			}

		}
	}

	return comp
}
