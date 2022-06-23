package report

import "strings"

var replacer *strings.Replacer

type HashAlgorithm string
type Package struct {
	PackageName             string            `json:"name,omitempty"`
	SPDXID                  string            `json:"SPDXID,omitempty"`
	PackageVersion          string            `json:"versionInfo,omitempty"`
	PackageSupplier         string            `json:"supplier,omitempty"`
	PackageDownloadLocation string            `json:"downloadLocation,omitempty"`
	FilesAnalyzed           bool              `json:"filesAnalyzed"`
	PackageChecksums        []PackageChecksum `json:"checksums"`
	PackageHomePage         string            `json:"homepage,omitempty"`
	PackageLicenseConcluded string            `json:"licenseConcluded,omitempty"`
	PackageLicenseDeclared  string            `json:"licenseDeclared,omitempty"`
	PackageCopyrightText    string            `json:"copyrightText,omitempty"`
	PackageLicenseComments  string            `json:"licenseComments,omitempty"`
	PackageComment          string            `json:"comment,omitempty"`
	RootPackage             bool              `json:"-"`
}

type Document struct {
	SPDXVersion             string                   `json:"spdxVersion,omitempty"`
	DataLicense             string                   `json:"dataLicense,omitempty"`
	SPDXID                  string                   `json:"SPDXID,omitempty"`
	DocumentName            string                   `json:"name,omitempty"`
	DocumentNamespace       string                   `json:"documentNamespace,omitempty"`
	CreationInfo            CreationInfo             `json:"creationInfo,omitempty"`
	Packages                []Package                `json:"packages,omitempty"`
	Relationships           []Relationship           `json:"relationships,omitempty"`
	ExtractedLicensingInfos []ExtractedLicensingInfo `json:"hasExtractedLicensingInfos,omitempty"`
}

type CreationInfo struct {
	Comment            string   `json:"comment,omitempty"`
	Created            string   `json:"created,omitempty"`
	Creators           []string `json:"creators,omitempty"`
	LicenceListVersion string   `json:"licenseListVersion,omitempty"`
}

type Relationship struct {
	SPDXElementID      string `json:"spdxElementId,omitempty"`
	RelatedSPDXElement string `json:"relatedSpdxElement,omitempty"`
	RelationshipType   string `json:"relationshipType,omitempty"`
}
type ExtractedLicensingInfo struct {
	LicenseID      string `json:"licenseId,omitempty"`
	ExtractedText  string `json:"extractedText,omitempty"`
	LicenseName    string `json:"name,omitempty"`
	LicenseComment string `json:"comment,omitempty"`
}
type PackageChecksum struct {
	Algorithm HashAlgorithm `json:"algorithm"`
	Value     string        `json:"checksumValue"`
}

const T = `SPDXVersion: {{ .SPDXVersion }}
DataLicense: {{ .DataLicense }}
SPDXID: {{ .SPDXID }}
DocumentName: {{ .DocumentName }}
DocumentNamespace: {{ .DocumentNamespace }}
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
PackageDownloadLocation: {{ .PackageDownloadLocation }}
FilesAnalyzed: {{ .FilesAnalyzed }}
{{- range .PackageChecksums }}
PackageChecksum: {{ .Algorithm }}: {{ .Value }}
{{- end }}
PackageHomePage: {{ .PackageHomePage }}
PackageLicenseConcluded: {{ .PackageLicenseConcluded }}
PackageLicenseDeclared: {{ .PackageLicenseDeclared }}
PackageCopyrightText: {{ .PackageCopyrightText }}
PackageLicenseComments: {{ .PackageLicenseComments }}
PackageComment: {{ .PackageComment }}
{{ end }}
{{- range .Relationships }}
Relationship: {{ .SPDXElementID }} {{ .RelationshipType }} {{ .RelatedSPDXElement }}
{{- end }}

{{- with .ExtractedLicensingInfos -}}
##### Non-standard license
{{ range . }}
LicenseID: {{ .LicenseID }}
ExtractedText: {{ .ExtractedText }}
LicenseName: {{ .LicenseName }}
LicenseComment: {{ .LicenseComment }}
{{- end -}}
{{- end -}}`
