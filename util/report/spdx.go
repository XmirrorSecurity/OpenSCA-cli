package report

import (
	"bytes"
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
	replacers := []string{"/", ".", "_", "-", `\`, "."}
	replacer = strings.NewReplacer(replacers...)
}
func Spdx(dep *model.DepTree, taskInfo TaskInfo) []byte {
	format(dep)
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
		logs.Warn(err)
	}
	return templateBuffer.Bytes()
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
	q := []*model.DepTree{root}
	for len(q) > 0 {
		n := q[0]
		q = append(q[1:], n.Children...)
		doc.Packages = append(doc.Packages, buildPkg(n))
	}
}

// 构建package
func buildPkg(dep *model.DepTree) Package {
	pkg := Package{
		PackageName:             dep.Name,
		SPDXID:                  "",
		PackageVersion:          dep.VersionStr,
		PackageSupplier:         dep.Vendor,
		PackageDownloadLocation: "NOASSERTION",
		FilesAnalyzed:           false,
		PackageChecksums:        []PackageChecksum{{}},
		PackageHomePage:         "NOASSERTION",
		PackageLicenseConcluded: "NOASSERTION",
		PackageLicenseDeclared:  "NOASSERTION",
		PackageCopyrightText:    "NOASSERTION",
		PackageLicenseComments:  "NOASSERTION",
		PackageComment:          "NOASSERTION",
		RootPackage:             len(dep.Children) > 0,
	}
	pkg.SPDXID = setPkgSPDXID(dep.Name, dep.VersionStr, pkg.RootPackage)
	nodePkg[dep] = pkg
	return pkg
}

// 初始化Document
func buildDocument(root *model.DepTree, taskInfo TaskInfo) *Document {
	return &Document{
		SPDXVersion:       "SPDX-2.2",
		DataLicense:       "CC0-1.0",
		SPDXID:            "SPDXRef-DOCUMENT",
		DocumentName:      path.Base(taskInfo.AppName),
		DocumentNamespace: "",
		CreationInfo: CreationInfo{
			Creators: []string{},
			Created:  time.Now().UTC().Format(time.RFC3339),
		},
		Packages:                []Package{},
		Relationships:           []Relationship{},
		ExtractedLicensingInfos: []ExtractedLicensingInfo{},
	}
}

// 设置package的SPDXID
func setPkgSPDXID(s, v string, flag bool) string {
	if flag {
		return fmt.Sprintf("SPDXRef-Package-%s", replacer.Replace(s))
	}
	return fmt.Sprintf("SPDXRef-Package-%s-%s", replacer.Replace(s), v)
}
