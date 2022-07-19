package report

import "strings"

var replacer *strings.Replacer

type HashAlgorithm string
type Package struct {
	PackageName             string `json:"name,omitempty" xml:"name,omitempty"`
	SPDXID                  string `json:"SPDXID,omitempty" xml:"SPDXID,omitempty"`
	PackageVersion          string `json:"versionInfo,omitempty" xml:"versionInfo,omitempty"`
	PackageSupplier         string `json:"supplier,omitempty" xml:"supplier,omitempty"`
	PackageDownloadLocation string `json:"downloadLocation,omitempty" xml:"downloadLocation,omitempty"`
	// FilesAnalyzed           bool              `json:"filesAnalyzed" xml:"filesAnalyzed"`
	// PackageChecksums        []PackageChecksum `json:"checksums,omitempty" xml:"checksums>checksum,omitempty"`
	PackageHomePage         string `json:"homepage,omitempty" xml:"homepage,omitempty"`
	PackageLicenseConcluded string `json:"licenseConcluded,omitempty" xml:"licenseConcluded,omitempty"`
	PackageLicenseDeclared  string `json:"licenseDeclared,omitempty" xml:"licenseDeclared,omitempty"`
	PackageCopyrightText    string `json:"copyrightText,omitempty" xml:"copyrightText,omitempty"`
	PackageLicenseComments  string `json:"licenseComments,omitempty" xml:"licenseComments,omitempty"`
	PackageComment          string `json:"comment,omitempty" xml:"comment,omitempty"`
	RootPackage             bool   `json:"-" xml:"-"`
}

type Document struct {
	SPDXVersion             string                   `json:"spdxVersion,omitempty" xml:"spdxVersion,omitempty"`
	DataLicense             string                   `json:"dataLicense,omitempty" xml:"dataLicense,omitempty"`
	SPDXID                  string                   `json:"SPDXID,omitempty" xml:"SPDXID,omitempty"`
	DocumentName            string                   `json:"name,omitempty" xml:"name,omitempty"`
	DocumentNamespace       string                   `json:"documentNamespace,omitempty" xml:"documentNamespace,omitempty"`
	CreationInfo            CreationInfo             `json:"creationInfo,omitempty" xml:"creationInfo,omitempty"`
	Packages                []Package                `json:"packages,omitempty" xml:"packages>package,omitempty"`
	Relationships           []Relationship           `json:"relationships,omitempty" xml:"relationships>relationship,omitempty"`
	ExtractedLicensingInfos []ExtractedLicensingInfo `json:"hasExtractedLicensingInfos,omitempty" xml:"hasExtractedLicensingInfos>ExtractedLicensingInfo,omitempty"`
}

type CreationInfo struct {
	Comment            string   `json:"comment,omitempty" xml:"comment,omitempty"`
	Created            string   `json:"created,omitempty" xml:"created,omitempty"`
	Creators           []string `json:"creators,omitempty" xml:"creators>creator,omitempty"`
	LicenceListVersion string   `json:"licenseListVersion,omitempty" xml:"licenseListVersion,omitempty"`
}

type Relationship struct {
	SPDXElementID      string `json:"spdxElementId,omitempty" xml:"spdxElementId,omitempty"`
	RelatedSPDXElement string `json:"relatedSpdxElement,omitempty" xml:"relatedSpdxElement,omitempty"`
	RelationshipType   string `json:"relationshipType,omitempty" xml:"relationshipType,omitempty"`
}
type ExtractedLicensingInfo struct {
	LicenseID      string `json:"licenseId,omitempty" xml:"licenseId,omitempty"`
	ExtractedText  string `json:"extractedText,omitempty" xml:"extractedText,omitempty"`
	LicenseName    string `json:"name,omitempty" xml:"name,omitempty"`
	LicenseComment string `json:"comment,omitempty" xml:"comment,omitempty"`
}
type PackageChecksum struct {
	Algorithm HashAlgorithm `json:"algorithm,omitempty" xml:"algorithm,omitempty"`
	Value     string        `json:"checksumValue,omitempty" xml:"checksumValue,omitempty"`
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

// {{- range .PackageChecksums }}
// PackageChecksum: {{ .Algorithm }}: {{ .Value }}
// {{- end }}

// FilesAnalyzed: {{ .FilesAnalyzed }}
