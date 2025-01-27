package model

import "time"

type DpSbomDocument struct {
	// 文档名称
	DocumentName string `json:"DocumentName"`
	// 文档版本
	DocumentVersion string `json:"DocumentVersion"`
	// 文档创建/更新时间 yyyy-MM-ddTHH:mm:ssTZD
	DocumentTime string `json:"DocumentTime"`
	// 文档格式
	BomFormat string `json:"BomFormat"`
	// 生成工具
	Tool string `json:"tool"`
	// sbom签名信息
	Hashes DpSbomHashes `json:"Hashes"`
	// 组件列表
	Packages []DpSbomPackage `json:"Packages"`
	// 依赖关系
	Dependencies []DpSbomDependencies `json:"Dependencies"`
}

type DpSbomPackage struct {
	Name    string `json:"ComponentName"`
	Version string `json:"ComponentVersion"`

	Identifier struct {
		Purl string `json:"PURL"`
	} `json:"ComponentIdentifier"`

	License []string `json:"License"`

	Author   []map[string]string `json:"Author"`
	Provider []map[string]string `json:"Provider"`
	Hash     DpSbomHash          `json:"ComponentHash"`

	// 组件信息更新时间 yyyy-MM-ddTHH:mm:ssTZD
	Timestamp string `json:"Timestamp"`
}

type DpSbomDependencies struct {
	Ref       string `json:"Ref"`
	DependsOn []struct {
		Target string `json:"Target"`
	} `json:"DependsOn"`
}

func newDependencies(ref string, dependsOn []string) DpSbomDependencies {
	deps := DpSbomDependencies{Ref: ref}
	deps.DependsOn = make([]struct {
		Target string "json:\"Target\""
	}, len(dependsOn))
	for i, d := range dependsOn {
		deps.DependsOn[i].Target = d
	}
	return deps
}

type DpSbomHashes struct {
	Algorithm   string `json:"Algorithm"`
	HashFile    string `json:"HashFile,omitempty"`
	DigitalFile string `json:"DigitalFile,omitempty"`
}

type DpSbomHash struct {
	Algorithm string `json:"Algorithm,omitempty"`
	Hash      string `json:"Hash,omitempty"`
}

func NewDpSbomDocument(name, creator string) *DpSbomDocument {
	version := "1.0.0"
	timestamp := time.Now().Format("2006-01-02T15:04:05MST")
	return &DpSbomDocument{
		DocumentName:    name,
		DocumentVersion: version,
		DocumentTime:    timestamp,
		BomFormat:       "DP-SBOM-1.0",
		Tool:            creator,
		Hashes: DpSbomHashes{
			Algorithm: "SHA-256",
			HashFile:  "sha256.txt",
		},
		Dependencies: []DpSbomDependencies{},
	}
}

func (doc *DpSbomDocument) AppendComponents(fn func(*DpSbomPackage)) {
	c := DpSbomPackage{}
	if fn != nil {
		fn(&c)
	}
	if c.Timestamp == "" {
		c.Timestamp = time.Now().Format("2006-01-02T15:04:05MST")
	}
	if c.Author == nil {
		c.Author = []map[string]string{}
	}
	if c.Provider == nil {
		c.Provider = []map[string]string{}
	}
	doc.Packages = append(doc.Packages, c)
}

func (doc *DpSbomDocument) AppendDependencies(parentId string, childrenIds []string) {
	if doc.Dependencies == nil {
		doc.Dependencies = []DpSbomDependencies{}
	}
	if len(childrenIds) > 0 {
		doc.Dependencies = append(doc.Dependencies, newDependencies(parentId, childrenIds))
	}
}
