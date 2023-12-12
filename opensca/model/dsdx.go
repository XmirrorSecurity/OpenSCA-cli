package model

import (
	"encoding/json"
	"fmt"
	"io"
	"text/template"
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
	Dependencies DsdxDependencies `json:"dependencies" xml:"-"`
}

type DsdxComponent struct {
	// DSDX-xxx
	ID       string   `json:"id" xml:"id"`
	Group    string   `json:"group,omitempty" xml:"group,omitempty"`
	Name     string   `json:"name" xml:"name"`
	Version  string   `json:"version" xml:"version"`
	Language string   `json:"language,omitempty" xml:"language,omitempty"`
	License  []string `json:"license,omitempty" xml:"license,omitempty"`
}

type DsdxDependencies map[string][]string

func NewDsdxDocument(name, creator string) *DsdxDocument {
	version := "DSDX-1.0"
	create := time.Now().Format("2006-01-02 15:04:05")
	id := fmt.Sprintf("DSDX-%s-%s-%s", name, version, create)
	return &DsdxDocument{
		Name:        name,
		Creator:     creator,
		ProjectName: name,
		CreateTime:  create,
		DSDXVersion: version,
		DSDXID:      id,
	}
}

func (doc *DsdxDocument) AppendComponents(id, group, name, version, language string, license []string) {
	if id == "" {
		id = fmt.Sprintf("DSDX-%s-%s-%s", group, name, version)
	}
	doc.Components = append(doc.Components, DsdxComponent{
		ID:       id,
		Group:    group,
		Name:     name,
		Version:  version,
		Language: language,
		License:  license,
	})
}

func (doc *DsdxDocument) AppendDependencies(parentId string, childrenIds []string) {
	if doc.Dependencies == nil {
		doc.Dependencies = DsdxDependencies{}
	}
	if len(childrenIds) > 0 {
		doc.Dependencies[parentId] = childrenIds
	}
}

func (doc *DsdxDocument) WriteDsdx(w io.Writer) error {
	tmpl, err := template.New("tagValue").Funcs(template.FuncMap{
		"tojson": func(o any) string {
			data, _ := json.Marshal(o)
			return string(data)
		},
	}).Parse(dsdxtpl)
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
ComponentID: {{ .ID }}
ComponentLanguage: {{ .Language }}
ComponentLicense: {{ .License|tojson }}
{{ end }}
Dependencies: {{ .Dependencies|tojson }}
`
