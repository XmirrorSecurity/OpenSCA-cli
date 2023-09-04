package model

import (
	"html/template"
	"io"
	"strings"

)

type SpdxDocument struct {
	SPDXVersion   string         `json:"spdxVersion,omitempty" xml:"spdxVersion,omitempty"`
	SPDXID        string         `json:"SPDXID,omitempty" xml:"SPDXID,omitempty"`
	DocumentName  string         `json:"name,omitempty" xml:"name,omitempty"`
	CreationInfo  CreationInfo   `json:"creationInfo,omitempty" xml:"creationInfo,omitempty"`
	Packages      []SpdxPackage  `json:"packages,omitempty" xml:"packages>package,omitempty"`
	Relationships []Relationship `json:"relationships,omitempty" xml:"relationships>relationship,omitempty"`
}

type SpdxPackage struct {
	SPDXID                  string `json:"SPDXID,omitempty" xml:"SPDXID,omitempty"`
	PackageName             string `json:"name,omitempty" xml:"name,omitempty"`
	PackageVersion          string `json:"versionInfo,omitempty" xml:"versionInfo,omitempty"`
	PackageSupplier         string `json:"supplier,omitempty" xml:"supplier,omitempty"`
	PackageLicenseConcluded string `json:"licenseConcluded,omitempty" xml:"licenseConcluded,omitempty"`
}

type CreationInfo struct {
	Created  string   `json:"created,omitempty" xml:"created,omitempty"`
	Creators []string `json:"creators,omitempty" xml:"creators>creator,omitempty"`
}

type Relationship struct {
	SPDXElementID      string `json:"spdxElementId,omitempty" xml:"spdxElementId,omitempty"`
	RelatedSPDXElement string `json:"relatedSpdxElement,omitempty" xml:"relatedSpdxElement,omitempty"`
	RelationshipType   string `json:"relationshipType,omitempty" xml:"relationshipType,omitempty"`
}

func NewSpdxDocument(name, created string) *SpdxDocument {
	return &SpdxDocument{
		SPDXVersion:  "SPDX-2.2",
		SPDXID:       "SPDXRef-DOCUMENT",
		DocumentName: name,
		CreationInfo: CreationInfo{Created: created, Creators: []string{"opensca-cli"}},
	}
}

func (doc *SpdxDocument) AddPackage(id, vendor, name, version string, lics []string) {
	doc.Packages = append(doc.Packages, SpdxPackage{
		SPDXID:                  "SPDXRef-Package-" + id,
		PackageName:             name,
		PackageVersion:          version,
		PackageSupplier:         vendor,
		PackageLicenseConcluded: strings.Join(lics, " OR "),
	})
}

func (doc *SpdxDocument) AddRelation(parentId, childId string) {
	doc.Relationships = append(doc.Relationships, Relationship{
		SPDXElementID:      "SPDXRef-Package-" + parentId,
		RelatedSPDXElement: "SPDXRef-Package-" + childId,
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

const spdxtpl = `SPDXVersion: {{ .SPDXVersion }}
SPDXID: {{ .SPDXID }}
DocumentName: {{ .DocumentName }}
Creator: {{ range .CreationInfo.Creators }}{{ . -}} {{ end }}
Created: {{ .CreationInfo.Created }}

{{ range .Packages }}
##### Package representing the {{.PackageName}}

PackageName: {{ .PackageName }}
SPDXID: {{ .SPDXID }}
{{ with .PackageVersion -}}
PackageVersion: {{ . }}
{{- end }}
PackageSupplier: {{ .PackageSupplier }}
PackageLicenseConcluded: {{ .PackageLicenseConcluded }}
{{ end }}
{{- range .Relationships }}
Relationship: {{ .SPDXElementID }} {{ .RelationshipType }} {{ .RelatedSPDXElement }}
{{- end }}`
