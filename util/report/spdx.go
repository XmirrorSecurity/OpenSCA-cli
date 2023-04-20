package report

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"path"
	"strings"
	"text/template"
	"time"
	"util/logs"
	"util/model"
)

// 记录节点名与pacakge的对应关系
var nodePkg = make(map[*model.DepTree]Package)

func init() {
	replacers := []string{"_", "-", "/", "."}
	replacer = strings.NewReplacer(replacers...)
}

func Spdx(dep *model.DepTree, taskInfo TaskInfo) []byte {
	doc := buildDocument(dep, taskInfo)
	addPkgToDoc(dep, doc)
	addRelation(dep, doc)
	tmpl := template.New("tagValue")
	tmpl, err := tmpl.Parse(T)

	if err != nil {
		logs.Warn(err)
	}
	templateBuffer := new(bytes.Buffer)
	err = tmpl.Execute(templateBuffer, doc)
	if err != nil {
		logs.Error(err)
	}
	return templateBuffer.Bytes()
}

func SpdxJson(dep *model.DepTree, taskInfo TaskInfo) []byte {
	doc := buildDocument(dep, taskInfo)
	addPkgToDoc(dep, doc)
	addRelation(dep, doc)
	type D struct {
		Document `json:"document"`
	}
	d := D{*doc}
	res, err := json.Marshal(d.Document)
	if err != nil {
		logs.Error(err)
	}
	return res
}

func SpdxXml(dep *model.DepTree, taskInfo TaskInfo) []byte {
	doc := buildDocument(dep, taskInfo)
	addPkgToDoc(dep, doc)
	addRelation(dep, doc)
	type D struct {
		Document `xml:"document"`
	}
	d := D{*doc}
	res, err := xml.Marshal(d.Document)
	if err != nil {
		logs.Error(err)
	}
	return res
}

// 为document添加relationship字段
func addRelation(dep *model.DepTree, doc *Document) {
	doc.Relationships = append(doc.Relationships, Relationship{
		SPDXElementID:      "SPDXRef-DOCUMENT",
		RelatedSPDXElement: doc.DocumentName,
		RelationshipType:   "DESCRIBES",
	})
	q := []*model.DepTree{dep}
	for len(q) > 0 {
		n := q[0]
		if pkg, ok := nodePkg[n]; ok {
			if !pkg.RootPackage {
				q = append(q[1:], n.Children...)
				continue
			}
			for _, sub := range n.Children {
				if subpkg, ok := nodePkg[sub]; ok {
					doc.Relationships = append(doc.Relationships, Relationship{
						SPDXElementID:      pkg.SPDXID,
						RelatedSPDXElement: subpkg.SPDXID,
						RelationshipType:   "DEPENDS_ON",
					})
				}
			}
		}
		q = append(q[1:], n.Children...)
	}
}

// 为document添加packages字段
func addPkgToDoc(root *model.DepTree, doc *Document) {
	if root.Name == "" {
		root.Name = doc.DocumentName
	}
	q := []*model.DepTree{}
	q = append(q, root.Children...)
	for len(q) > 0 {
		n := q[0]
		q = append(q[1:], n.Children...)
		doc.Packages = append(doc.Packages, buildPkg(n))
	}
}

// 构建package
func buildPkg(dep *model.DepTree) Package {
	pkg := Package{
		PackageName:             setpkgName(dep),
		SPDXID:                  "",
		PackageVersion:          setPkgVer(dep),
		PackageSupplier:         setPkgSup(dep),
		PackageDownloadLocation: setPkgDownloadLoc(dep),
		// FilesAnalyzed:           false,
		//PackageChecksums:        nil,
		PackageHomePage:         setHomePage(dep),
		PackageLicenseConcluded: setPkgLicenseCon(dep),
		PackageLicenseDeclared:  setPkgLicenseDec(dep),
		PackageCopyrightText:    setCopyrightCont(dep),
		PackageLicenseComments:  setPkgLicenseComments(dep),
		PackageComment:          setPkgComments(dep),
		RootPackage:             isParent(dep),
	}
	pkg.SPDXID = setPkgSPDXID(path.Base(dep.Name), dep.VersionStr)
	nodePkg[dep] = pkg
	return pkg
}

// 初始化Document
func buildDocument(root *model.DepTree, taskInfo TaskInfo) *Document {
	return &Document{
		SPDXVersion:       "SPDX-2.2",
		DataLicense:       "",
		SPDXID:            "SPDXRef-DOCUMENT",
		DocumentName:      path.Base(taskInfo.AppName),
		DocumentNamespace: "",
		CreationInfo: CreationInfo{
			Creators: []string{"OpenSCA-Cli"},
			Created:  time.Now().Format("2006-01-02 15:04:05"),
		},
		Packages:                []Package{},
		Relationships:           []Relationship{},
		ExtractedLicensingInfos: []ExtractedLicensingInfo{},
	}
}

func setPkgSPDXID(s, v string) string {
	if v == "" {
		return fmt.Sprintf("SPDXRef-Package-%s", replacer.Replace(s))
	}
	return fmt.Sprintf("SPDXRef-Package-%s-%s", replacer.Replace(s), v)
}
func setpkgName(dep *model.DepTree) string {
	if dep.Name != "" {
		return dep.Name
	}
	return ""
}
func setPkgVer(dep *model.DepTree) string {
	if dep.VersionStr != "" {
		return dep.VersionStr
	}
	return "NOASSERTION"
}
func setPkgSup(dep *model.DepTree) string {
	if dep.Vendor != "" {
		return dep.Vendor
	}
	return "NOASSERTION"
}
func setPkgDownloadLoc(dep *model.DepTree) string {
	if dep.DownloadLocation != "" {
		return dep.DownloadLocation
	}
	return "NOASSERTION"
}
func setHomePage(dep *model.DepTree) string {
	if dep.HomePage != "" {
		return dep.HomePage
	}
	return "NOASSERTION"
}
func setPkgLicenseCon(dep *model.DepTree) string {
	if len(dep.Licenses) > 0 {
		lic := ""
		for _, v := range dep.Licenses {
			if lic == "" {
				lic = v
				continue
			}
			lic = lic + " OR " + v
		}
		return lic
	}
	return "NOASSERTION"
}
func setPkgLicenseDec(dep *model.DepTree) string {
	return "NOASSERTION"
}
func setCopyrightCont(dep *model.DepTree) string {
	if dep.CopyrightText != "" {
		return dep.CopyrightText
	}
	return "NOASSERTION"
}
func setPkgLicenseComments(dep *model.DepTree) string {
	return "NOASSERTION"
}
func setPkgComments(dep *model.DepTree) string {
	return "NOASSERTION"
}
func isParent(dep *model.DepTree) bool {
	return len(dep.Children) > 0
}
