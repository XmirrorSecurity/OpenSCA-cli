package origin

import (
	"strings"
	"util/enum/language"
	"util/model"
)

type VulnInfo struct {
	Vendor   string `json:"vendor" gorm:"column:vendor"`
	Product  string `json:"product" gorm:"column:product"`
	Version  string `json:"version" gorm:"column:version"`
	Language string `json:"language" gorm:"column:language"`
	*model.Vuln
}

type BaseOrigin struct {
	// origin data
	// map[language]map[component_name][]VulnInfo
	data  map[string]map[string][]VulnInfo
	idSet map[string]bool
}

func NewBaseOrigin() *BaseOrigin {
	return &BaseOrigin{
		data:  map[string]map[string][]VulnInfo{},
		idSet: map[string]bool{},
	}
}

func (o *BaseOrigin) LoadDataOrigin(data ...VulnInfo) {
	if o == nil {
		return
	}
	for _, info := range data {
		if info.Vuln == nil {
			continue
		}
		if o.idSet[info.Id] {
			continue
		}
		o.idSet[info.Id] = true
		if info.Description != "" {
			info.DescriptionEn = ""
		}
		name := strings.ToLower(info.Product)
		language := strings.ToLower(info.Language)
		if _, ok := o.data[language]; !ok {
			o.data[language] = map[string][]VulnInfo{}
		}
		vulns := o.data[language]
		vulns[name] = append(vulns[name], info)
	}
}

func (o *BaseOrigin) SearchVuln(deps []model.Dependency) (vulns [][]*model.Vuln) {
	if o == nil || o.data == nil {
		return nil
	}
	vulns = make([][]*model.Vuln, len(deps))
	for i, dep := range deps {
		vulns[i] = []*model.Vuln{}
		if vs, ok := o.data[dep.Language.Vuln()][strings.ToLower(dep.Name)]; ok {
			for _, v := range vs {
				if strings.EqualFold(dep.Language.Vuln(), language.Java.Vuln()) && !strings.EqualFold(v.Vendor, dep.Vendor) {
					continue
				}
				if model.InRangeInterval(dep.Version, v.Version) {
					vulns[i] = append(vulns[i], v.Vuln)
				}
			}
		}
	}
	return
}
