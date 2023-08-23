package detail

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/xmirrorsecurity/opensca-cli/opensca/model"
	"github.com/xmirrorsecurity/opensca-cli/util/args"
	"github.com/xmirrorsecurity/opensca-cli/util/logs"
)

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

func vulnLanguageKey(language string) string {
	return language
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

type License struct {
	ShortName string `json:"name"`
}

type DepDetailGraph struct {
	Dep
	Path            string            `json:"path,omitempty" xml:"path,omitempty"`
	Licenses        []License         `json:"licenses,omitempty" xml:"licenses,omitempty"`
	Vulnerabilities []*Vuln           `json:"vulnerabilities,omitempty" xml:"vulnerabilities,omitempty" `
	Children        []*DepDetailGraph `json:"children,omitempty" xml:"children,omitempty"`
}

func (d *DepDetailGraph) Update(dep *model.DepGraph) {
	d.Name = dep.Name
	d.Vendor = dep.Vendor
	d.Version = dep.Version
	d.Language = string(dep.Language)
	d.Path = dep.Path
	for _, lic := range dep.Licenses {
		d.Licenses = append(d.Licenses, License{ShortName: lic})
	}
}

// SearchDetail 查找组件详情:漏洞/许可证
func SearchDetail(depRoot *model.DepGraph) (detailRoot *DepDetailGraph, err error) {

	var details []*DepDetailGraph
	var ds []Dep

	detailRoot = &DepDetailGraph{}
	depRoot.Expand = detailRoot
	depRoot.ForEachOnce(func(n *model.DepGraph) bool {
		detail := n.Expand.(*DepDetailGraph)
		detail.Update(n)
		for c := range n.Children {
			cd := &DepDetailGraph{}
			c.Expand = cd
			detail.Children = append(detail.Children, cd)
		}
		details = append(details, detail)
		ds = append(ds, detail.Dep)
		n.Expand = nil
		return true
	})

	serverVulns := [][]*Vuln{}
	localVulns := GetOrigin().SearchVuln(ds)
	if args.Config.Url != "" && args.Config.Token != "" {
		// vulnerability
		serverVulns, err = GetServerVuln(ds)
		// license
		serverLicenses, _ := GetServerLicense(ds)
		for i, lics := range serverLicenses {
			details[i].Licenses = append(details[i].Licenses, lics...)
		}
	} else if len(localVulns) == 0 {
		if args.Config.Url == "" && args.Config.Token != "" {
			err = errors.New("url is null")
		} else if args.Config.Url != "" && args.Config.Token == "" {
			err = errors.New("token is null")
		}
	}

	for i, detail := range details {
		// 合并本地和云端库搜索的漏洞
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
		lanKey := vulnLanguageKey(dep.Language)
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
