package model

import (
	"io"
	"strings"
	"text/template"
	"time"
)

type SpdxDocument struct {
	Namespace     string         `json:"documentNamespace" xml:"documentNamespace"`
	SPDXVersion   string         `json:"spdxVersion" xml:"spdxVersion"`
	SPDXID        string         `json:"SPDXID" xml:"SPDXID"`
	DocumentName  string         `json:"name" xml:"name"`
	CreationInfo  CreationInfo   `json:"creationInfo" xml:"creationInfo"`
	Packages      []SpdxPackage  `json:"packages" xml:"packages"`
	Relationships []Relationship `json:"relationships" xml:"relationships"`
}

type SpdxPackage struct {
	SPDXID       string        `json:"SPDXID" xml:"SPDXID"`
	Name         string        `json:"name" xml:"name"`
	Version      string        `json:"versionInfo,omitempty" xml:"versionInfo,omitempty"`
	Supplier     string        `json:"supplier,omitempty" xml:"supplier,omitempty"`
	ExternalRefs []ExternalRef `json:"externalRefs" xml:"externalRefs"`
	// 从文件中解析的许可证名称不符合spdx规范
	LicenseConcluded string `json:"-" xml:"-"`
}

type CreationInfo struct {
	Created  string   `json:"created,omitempty" xml:"created,omitempty"`
	Creators []string `json:"creators,omitempty" xml:"creators,omitempty"`
}

type Relationship struct {
	SPDXElementID      string `json:"spdxElementId" xml:"spdxElementId"`
	RelatedSPDXElement string `json:"relatedSpdxElement" xml:"relatedSpdxElement"`
	RelationshipType   string `json:"relationshipType" xml:"relationshipType"`
}

type ExternalRef struct {
	ReferenceCategory string `json:"referenceCategory" xml:"referenceCategory"`
	ReferenceLocator  string `json:"referenceLocator" xml:"referenceLocator"`
	ReferenceType     string `json:"referenceType" xml:"referenceType"`
}

func NewSpdxDocument(name string) *SpdxDocument {
	return &SpdxDocument{
		SPDXVersion:  "SPDX-2.2",
		SPDXID:       "SPDXRef-DOCUMENT",
		Namespace:    "ftp://spdx",
		DocumentName: name,
		CreationInfo: CreationInfo{Created: time.Now().Format("2006-01-02T15:04:05Z"), Creators: []string{"Tool: opensca-cli"}},
	}
}

func (doc *SpdxDocument) AddPackage(id, vendor, name, version string, language Language, lics []string) {
	purlRef := ExternalRef{
		ReferenceCategory: "PACKAGE-MANAGER",
		ReferenceType:     "purl",
		ReferenceLocator:  Purl(vendor, name, version, language),
	}
	doc.Packages = append(doc.Packages, SpdxPackage{
		SPDXID:           "SPDXRef-" + id,
		Name:             assert(name),
		Version:          assert(version),
		Supplier:         assert("Organization: " + vendor),
		LicenseConcluded: assert(strings.Join(lics, " OR ")),
		ExternalRefs:     []ExternalRef{purlRef},
	})
}

func (doc *SpdxDocument) AddRelation(parentId, childId string) {
	doc.Relationships = append(doc.Relationships, Relationship{
		SPDXElementID:      "SPDXRef-" + parentId,
		RelatedSPDXElement: "SPDXRef-" + childId,
		RelationshipType:   "DEPENDS_ON",
	})
}

func (doc *SpdxDocument) WriteSpdx(w io.Writer) error {
	tmpl, err := template.New("tagValue").Parse(spdxtpl)
	if err != nil {
		return err
	}
	return tmpl.Execute(w, doc)
}

func assert(s string) string {
	if s != "" {
		return s
	}
	return "NOASSERTION"
}

const spdxtpl = `SPDXVersion: {{ .SPDXVersion }}
SPDXID: {{ .SPDXID }}
DocumentName: {{ .DocumentName }}
DocumentNamespace: {{ .Namespace }}
Created: {{ .CreationInfo.Created }}
{{ range .CreationInfo.Creators -}}
Creator: {{ . }}
{{ end }}
{{- range .Packages }}
PackageName: {{ .Name }}
SPDXID: {{ .SPDXID }}
PackageVersion: {{ .Version }}
PackageSupplier: {{ .Supplier }}
{{- range .ExternalRefs	}}
ExternalRef: {{ .ReferenceCategory }} {{ .ReferenceType }} {{ .ReferenceLocator }}
{{- end }}
# PackageLicenseConcluded: {{ .LicenseConcluded }}
{{ end }}
{{- range .Relationships }}
Relationship: {{ .SPDXElementID }} {{ .RelationshipType }} {{ .RelatedSPDXElement }}
{{- end }}`
