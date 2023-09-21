package detail

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/xmirrorsecurity/opensca-cli/cmd/config"
	"github.com/xmirrorsecurity/opensca-cli/opensca/logs"
	"github.com/xmirrorsecurity/opensca-cli/opensca/model"
)

type DepDetailGraph struct {
	Dep
	ID                      string            `json:"id" xml:"id"`
	Develop                 bool              `json:"dev" xml:"dev"`
	Direct                  bool              `json:"direct" xml:"direct"`
	Paths                   []string          `json:"paths,omitempty" xml:"paths,omitempty"`
	Licenses                []License         `json:"licenses,omitempty" xml:"licenses,omitempty"`
	Vulnerabilities         []*Vuln           `json:"vulnerabilities,omitempty" xml:"vulnerabilities,omitempty" `
	Children                []*DepDetailGraph `json:"children,omitempty" xml:"children,omitempty"`
	Parent                  *DepDetailGraph   `json:"-" xml:"-"`
	IndirectVulnerabilities int               `json:"indirect_vulnerabilities,omitempty" xml:"indirect_vulnerabilities,omitempty" `
	Expand                  any               `json:"-" xml:"-"`
}

var (
	latestTime int64
	count      int64
	idMutex    sync.Mutex
)

// ID 生成一个本地唯一的id
func ID() string {
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
	return fmt.Sprint(res)
}

func NewDepDetailGraph(dep *model.DepGraph) *DepDetailGraph {
	detail := &DepDetailGraph{ID: ID()}
	detail.Update(dep)
	dep.Expand = detail
	dep.ForEachNode(func(p, n *model.DepGraph) bool {
		if p == nil || p.Expand == nil {
			return true
		}
		parent := p.Expand.(*DepDetailGraph)
		child := &DepDetailGraph{ID: ID(), Parent: parent}
		child.Update(n)
		n.Expand = child
		parent.Children = append(parent.Children, child)
		return true
	})
	return detail
}

func (d *DepDetailGraph) Update(dep *model.DepGraph) {
	d.Name = dep.Name
	d.Vendor = dep.Vendor
	d.Version = dep.Version
	d.Language = string(dep.Language)
	d.Paths = append(d.Paths, dep.Path)
	d.Direct = dep.Direct
	d.Develop = dep.Develop
	for _, lic := range dep.Licenses {
		d.Licenses = append(d.Licenses, License{ShortName: lic})
	}
}

func (d *DepDetailGraph) ForEach(do func(n *DepDetailGraph) bool) {
	if d == nil {
		return
	}
	q := []*DepDetailGraph{d}
	for len(q) > 0 {
		n := q[0]
		q = q[1:]
		if do(n) {
			q = append(q, n.Children...)
		}
	}
}

func (d *DepDetailGraph) RemoveDedup() {
	// map[key]
	depSet := map[string]*DepDetailGraph{}
	d.ForEach(func(n *DepDetailGraph) bool {
		dep, ok := depSet[n.Key()]
		if !ok {
			depSet[n.Key()] = n
			return true
		}
		dep.Paths = append(dep.Paths, n.Paths...)
		for i, c := range dep.Parent.Children {
			if c.ID == n.ID {
				dep.Parent.Children = append(dep.Parent.Children[:i], dep.Parent.Children[i+1:]...)
				break
			}
		}
		dep.Children = append(dep.Children, n.Children...)
		return true
	})
}

func (d *DepDetailGraph) RemoveDev() {
	d.ForEach(func(n *DepDetailGraph) bool {
		if !n.Develop {
			return true
		}
		if n.Parent == nil {
			return false
		}
		for i, c := range n.Parent.Children {
			if c.ID == n.ID {
				n.Parent.Children = append(n.Parent.Children[:i], n.Parent.Children[i+1:]...)
				break
			}
		}
		return false
	})
}

func (dep *DepDetailGraph) Purl() string {
	var purlMap = map[model.Language]string{
		model.Lan_Rust:       "cargo",
		model.Lan_Php:        "composer",
		model.Lan_Ruby:       "gem",
		model.Lan_Golang:     "golang",
		model.Lan_Java:       "maven",
		model.Lan_JavaScript: "npm",
		model.Lan_Python:     "pypi",
	}
	group := ""
	if g, ok := purlMap[model.Language(dep.Language)]; ok {
		group = g
	}
	if dep.Vendor == "" {
		return fmt.Sprintf("pkg:%s/%s@%s", group, dep.Name, dep.Version)
	}
	return fmt.Sprintf("pkg:%s/%s/%s@%s", group, dep.Vendor, dep.Name, dep.Version)
}

// Vuln 组件漏洞
type Vuln struct {
	Name            string `json:"name,omitempty" gorm:"column:name"`
	Id              string `json:"id" gorm:"column:id"`
	Cve             string `json:"cve_id,omitempty" gorm:"column:cve_id"`
	Cnnvd           string `json:"cnnvd_id,omitempty" gorm:"column:cnnvd_id"`
	Cnvd            string `json:"cnvd_id,omitempty" gorm:"column:cnvd_id"`
	Cwe             string `json:"cwe_id,omitempty" gorm:"column:cwe_id"`
	Description     string `json:"description,omitempty" gorm:"column:description"`
	DescriptionEn   string `json:"description_en,omitempty" gorm:"-"`
	Suggestion      string `json:"suggestion,omitempty" gorm:"column:suggestion"`
	AttackType      string `json:"attack_type,omitempty" gorm:"column:attack_type"`
	ReleaseDate     string `json:"release_date,omitempty" gorm:"column:release_date"`
	SecurityLevelId int    `json:"security_level_id" gorm:"column:security_level_id"`
	ExploitLevelId  int    `json:"exploit_level_id" gorm:"column:exploit_level_id"`
}

func vulnLanguageKey(language model.Language) string {
	switch language {
	case model.Lan_Java:
		return "java"
	case model.Lan_JavaScript:
		return "js"
	case model.Lan_Php:
		return "php"
	case model.Lan_Python:
		return "python"
	case model.Lan_Golang:
		return "golang"
	case model.Lan_Ruby:
		return "ruby"
	case model.Lan_Rust:
		return "rust"
	default:
		return ""
	}
}

type Dep struct {
	// 厂商
	Vendor string `json:"vendor"`
	// 名称
	Name string `json:"name"`
	// 版本号
	Version string `json:"version"`
	// 语言
	Language string `json:"language"`
}

func (d Dep) Key() string {
	return fmt.Sprintf("%s:%s:%s:%s", d.Vendor, d.Name, d.Version, d.Language)
}

type License struct {
	ShortName string `json:"name"`
}

// SearchDetail 查找组件详情:漏洞/许可证
func SearchDetail(detailRoot *DepDetailGraph) (err error) {

	var details []*DepDetailGraph
	var ds []Dep

	detailRoot.ForEach(func(n *DepDetailGraph) bool {
		details = append(details, n)
		ds = append(ds, n.Dep)
		return true
	})

	serverVulns := [][]*Vuln{}
	localVulns := GetOrigin().SearchVuln(ds)

	c := config.Conf().Origin
	if c.Url != "" && c.Token != "" {
		// vulnerability
		serverVulns, err = GetServerVuln(ds)
		// license
		serverLicenses, _ := GetServerLicense(ds)
		for i, lics := range serverLicenses {
			details[i].Licenses = append(details[i].Licenses, lics...)
		}
	} else if len(localVulns) == 0 {
		if c.Url == "" && c.Token != "" {
			err = errors.New("url is null")
		} else if c.Url != "" && c.Token == "" {
			err = errors.New("token is null")
		}
	}

	// 合并本地和云端库搜索的漏洞
	for i, detail := range details {
		exist := map[string]struct{}{}
		if len(localVulns) != 0 {
			for _, vuln := range localVulns[i] {
				if vuln.Id == "" {
					continue
				}
				if _, ok := exist[vuln.Id]; !ok {
					exist[vuln.Id] = struct{}{}
					detail.Vulnerabilities = append(detail.Vulnerabilities, vuln)
				}
			}
		}
		if len(serverVulns) != 0 {
			for _, vuln := range serverVulns[i] {
				if vuln.Id == "" {
					continue
				}
				if _, ok := exist[vuln.Id]; !ok {
					exist[vuln.Id] = struct{}{}
					detail.Vulnerabilities = append(detail.Vulnerabilities, vuln)
				}
			}
		}
	}

	// 统计关联/间接漏洞
	var deps []*DepDetailGraph
	detailRoot.ForEach(func(n *DepDetailGraph) bool {
		deps = append(deps, n)
		return true
	})
	indirect := map[string]map[string]struct{}{}
	for i := len(deps) - 1; i >= 0; i-- {
		dep := deps[i]
		// 记录当前依赖的关联漏洞
		m := map[string]struct{}{}
		for _, v := range dep.Vulnerabilities {
			m[v.Id] = struct{}{}
		}
		for _, c := range dep.Children {
			for id := range indirect[c.ID] {
				m[id] = struct{}{}
			}
			delete(indirect, c.ID)
		}
		dep.IndirectVulnerabilities = len(m)
		indirect[dep.ID] = m
	}

	return
}

// GetServerLicense 从云服务获取许可证
func GetServerLicense(deps []Dep) (lics [][]License, err error) {
	lics = [][]License{}
	data, err := json.Marshal(deps)
	if err != nil {
		logs.Error(err)
		return
	}
	data, err = Detect("license", data)
	if err != nil {
		fmt.Printf("\n%s", err.Error())
		return lics, err
	}
	if len(data) > 0 {
		err = json.Unmarshal(data, &lics)
		if err != nil {
			logs.Error(err)
		}
	}
	return
}

// GetServerVuln 从云服务获取漏洞
func GetServerVuln(deps []Dep) (vulns [][]*Vuln, err error) {
	vulns = [][]*Vuln{}
	data, err := json.Marshal(deps)
	if err != nil {
		logs.Error(err)
		return
	}
	data, err = Detect("vuln", data)
	if err != nil {
		fmt.Printf("\n%s", err.Error())
		return vulns, err
	}
	if len(data) > 0 {
		err = json.Unmarshal(data, &vulns)
		if err != nil {
			logs.Error(err)
		}
	}
	return
}

func (o *BaseOrigin) SearchVuln(deps []Dep) (vulns [][]*Vuln) {
	if o == nil || o.data == nil {
		return nil
	}
	vulns = make([][]*Vuln, len(deps))
	for i, dep := range deps {
		vulns[i] = []*Vuln{}
		lanKey := vulnLanguageKey(model.Language(dep.Language))
		if vs, ok := o.data[lanKey][strings.ToLower(dep.Name)]; ok {
			curVer := newVersion(dep.Version)
			for _, v := range vs {
				if strings.EqualFold(lanKey, "java") && !strings.EqualFold(v.Vendor, dep.Vendor) {
					continue
				}
				if inRangeInterval(curVer, v.Version) {
					vulns[i] = append(vulns[i], v.Vuln)
				}
			}
		}
	}
	return
}
