package model

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"time"
)

type DsdxDocument struct {
	// 文档名称
	Name string `json:"name" xml:"name"`
	// 创作者
	Creator string `json:"creator" xml:"creator"`
	// dsdx版本 DSDX-1.0
	DSDXVersion string `json:"dsdx_version" xml:"dsdx_version"`
	// 文档创建时间 yyyy-MM-dd HH:mm:ss
	CreateTime string `json:"create_time" xml:"create_time"`
	// dsdx文档标识 自动生成 DSDX-${Name}-${Version}-${CreateTime}
	DSDXID string `json:"dsdx_id" xml:"dsdx_id"`
	// 项目名称
	ProjectName string `json:"project_naem" xml:"project_naem"`
	// 组件列表
	Components []DsdxComponent `json:"components" xml:"components"`
	// 依赖关系
	Dependencies DsdxDependencies `json:"dependencies" xml:"dependencies"`
	// 自动生成
	DependenciesJson string `json:"-" xml:"-"`
}

type DsdxComponent struct {
	// DSDX-xxx
	ID       string `json:"id" xml:"id"`
	Group    string `json:"group" xml:"group"`
	Name     string `json:"name" xml:"name"`
	Version  string `json:"version" xml:"version"`
	Language string `json:"language" xml:"language"`
	// json list
	License []string `json:"license" xml:"license"`
	// 自动生成
	LicenseJson string `json:"-" xml:"-"`
}

type DsdxDependencies map[string][]string

func NewDsdxDocument(name, creator, project string) *DsdxDocument {
	create := time.Now().Format("2006-01-02 15:04:05")
	version := "DSDX-1.0"
	id := fmt.Sprintf("DSDX-%s-%s-%s", name, version, create)
	return &DsdxDocument{
		Name:        name,
		Creator:     creator,
		ProjectName: project,
		CreateTime:  create,
		DSDXVersion: version,
		DSDXID:      id,
	}
}

func (doc *DsdxDocument) AppendComponents(id, group, name, version, language string, license []string) {
	if id == "" {
		id = fmt.Sprintf("DSDX-%s-%s-%s", group, name, version)
	}
	lic, _ := json.Marshal(license)
	doc.Components = append(doc.Components, DsdxComponent{
		ID:          id,
		Group:       group,
		Name:        name,
		Version:     version,
		Language:    language,
		License:     license,
		LicenseJson: string(lic),
	})
}

func (doc *DsdxDocument) AppendDependencies(parentId string, childrenIds []string) {
	doc.Dependencies[parentId] = childrenIds
}

func (doc *DsdxDocument) WriteDsdx(w io.Writer) error {
	depJson, _ := json.Marshal(doc.Dependencies)
	doc.DependenciesJson = string(depJson)
	tmpl, err := template.New("tagValue").Parse(dsdxtpl)
	if err != nil {
		return err
	}
	return tmpl.Execute(w, doc)
}

const dsdxtpl = `Name: {{ .Name }}
DSDXVersion: {{ .DSDXVersion }}
DSDXID: {{ .DSDXID }}
Creator: {{ .Creator }}
CreateTime: {{ .CreateTime }}
ProjectName: {{ .ProjectName }}

{{ range .Components }}

ComponentName: {{ .Name }}
ComponentGroup: {{ .Group }}
ComponentVersion: {{ .Version }}
ComponentLanguage: {{ .Language }}
ComponentLicenses: {{ .LicenseJson }}
ComponentID: {{ .ID }}

{{ end }}

Dependencies: {{ .DependenciesJson }}
`
