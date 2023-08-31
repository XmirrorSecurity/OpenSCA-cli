package format

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"strings"
	"text/template"

	"github.com/xmirrorsecurity/opensca-cli/cmd/detail"
	"github.com/xmirrorsecurity/opensca-cli/opensca/logs"
)

func Spdx(report Report, out string) {
	tmpl, err := template.New("tagValue").Parse(T)
	if err != nil {
		logs.Warn(err)
		return
	}
	outWrite(out, func(w io.Writer) {
		err = tmpl.Execute(w, spdxDoc(report))
		if err != nil {
			logs.Warn(err)
		}
	})
}

func SpdxJson(report Report, out string) {
	outWrite(out, func(w io.Writer) {
		json.NewEncoder(w).Encode(spdxDoc(report))
	})
}

func SpdxXml(report Report, out string) {
	outWrite(out, func(w io.Writer) {
		xml.NewEncoder(w).Encode(spdxDoc(report))
	})
}

func spdxDoc(report Report) SpdxDocument {

	doc := NewSpdxDocument(report.AppName, report.EndTime)

	report.DepDetailGraph.ForEach(func(n *detail.DepDetailGraph) bool {
		doc.AddPackage(n)
		for _, c := range n.Children {
			doc.AddRelation(n, c)
		}
		return true
	})

	return doc
}

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

func NewSpdxDocument(name, created string) SpdxDocument {
	return SpdxDocument{
		SPDXVersion:  "SPDX-2.2",
		SPDXID:       "SPDXRef-DOCUMENT",
		DocumentName: name,
		CreationInfo: CreationInfo{Created: created, Creators: []string{"opensca-cli"}},
	}
}

func (doc *SpdxDocument) AddPackage(dep *detail.DepDetailGraph) {
	lics := []string{}
	for _, lic := range dep.Licenses {
		lics = append(lics, lic.ShortName)
	}
	doc.Packages = append(doc.Packages, SpdxPackage{
		SPDXID:                  "SPDXRef-Package-" + dep.ID,
		PackageName:             dep.Name,
		PackageVersion:          dep.Version,
		PackageSupplier:         dep.Vendor,
		PackageLicenseConcluded: strings.Join(lics, " OR "),
	})
}

func (doc *SpdxDocument) AddRelation(parent, child *detail.DepDetailGraph) {
	doc.Relationships = append(doc.Relationships, Relationship{
		SPDXElementID:      "SPDXRef-Package-" + parent.ID,
		RelatedSPDXElement: "SPDXRef-Package-" + child.ID,
		RelationshipType:   "DEPENDS_ON",
	})
}

const T = `SPDXVersion: {{ .SPDXVersion }}
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
