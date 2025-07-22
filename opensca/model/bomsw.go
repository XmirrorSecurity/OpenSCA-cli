package model

import (
	"time"
)

type BomSWDocument struct {
	Basic    swBasicInfo               `json:"documentBasicInfo"`
	Software swSoftwareCompositionInfo `json:"softwareCompositionInfo"`
}

type swBasicInfo struct {
	// 文档名称
	DocumentName string `json:"documentName"`
	// 文档版本
	DocumentVersion string `json:"documentVersion"`
	// 文档创建/更新时间 yyyy-MM-ddTHH:mm:ssTZD
	DocumentTime string `json:"timestamp"`
	// 文档格式
	SbomFormat string `json:"sbomFormat"`
	// 生成工具
	ToolInfo string `json:"toolInfo"`
	// bom作者
	SbomAuthor string `json:"sbomAuthor"`
	// 文档作者注释
	SbomAuthorComments string `json:"sbomAuthorComments"`
	// 文档注释
	SbomComments string `json:"sbomComments"`
	// 文档类型
	SbomType string `json:"sbomType"`
}

type swSoftwareCompositionInfo struct {
	// 组件列表
	Components []BomSWComponent `json:"components"`
	// 依赖关系
	Dependencies []swDependencies `json:"dependencies"`
}

type BomSWComponent struct {
	Author   map[string]string `json:"componentAuthor"`
	Provider map[string]string `json:"componentProvider"`
	Name     string            `json:"componentName"`
	Version  string            `json:"componentVersion"`
	// map[hash算法]hash值
	HashValue []swChecksumValue `json:"componentHashValue"`
	ID        string            `json:"componentId"`
	License   []string          `json:"license"`
	// 组件信息更新时间 yyyy-MM-ddTHH:mm:ssTZD
	Timestamp string `json:"componentTimestamp"`
}

type swChecksumValue struct {
	Algorithm string `json:"algorithm"`
	Value     string `json:"hashValue"`
}

type swDependencies struct {
	Ref       string `json:"ref"`
	DependsOn []struct {
		Ref string `json:"ref"`
	} `json:"dependsOn"`
}

func newDependencies(ref string, dependsOn []string) swDependencies {
	deps := swDependencies{Ref: ref}
	deps.DependsOn = make([]struct {
		Ref string `json:"ref"`
	}, len(dependsOn))
	for i, d := range dependsOn {
		deps.DependsOn[i].Ref = d
	}
	return deps
}

func NewBomSWDocument(name, creator string) *BomSWDocument {
	version := "1.0.0"
	timestamp := time.Now().Format("2006-01-02T15:04:05MST")
	return &BomSWDocument{
		Basic: swBasicInfo{
			DocumentName:       name,
			DocumentVersion:    version,
			DocumentTime:       timestamp,
			SbomFormat:         "BOM-SW 1.0",
			ToolInfo:           creator,
			SbomAuthor:         "",
			SbomAuthorComments: "",
			SbomComments:       "",
			SbomType:           "analyzed",
		},
		Software: swSoftwareCompositionInfo{
			Dependencies: []swDependencies{},
		},
	}
}

func (doc *BomSWDocument) AppendComponents(fn func(*BomSWComponent)) {
	c := BomSWComponent{
		Author: map[string]string{
			"name": "NONE",
		},
		Provider: map[string]string{
			"shortName": "NONE",
			"fullName":  "NONE",
		},
		HashValue: []swChecksumValue{},
		License:   []string{},
	}
	if fn != nil {
		fn(&c)
	}
	if c.Timestamp == "" {
		c.Timestamp = time.Now().Format("2006-01-02T15:04:05MST")
	}
	doc.Software.Components = append(doc.Software.Components, c)
}

func (doc *BomSWDocument) AppendDependencies(parentId string, childrenIds []string) {
	if doc.Software.Dependencies == nil {
		doc.Software.Dependencies = []swDependencies{}
	}
	if len(childrenIds) > 0 {
		doc.Software.Dependencies = append(doc.Software.Dependencies, newDependencies(parentId, childrenIds))
	}
}
