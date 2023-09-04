package sbom

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"strings"

	"github.com/xmirrorsecurity/opensca-cli/opensca/model"
)

func ParseSpdx(f *model.File) *model.DepGraph {

	depRelation := map[string][]string{}
	depIdMap := map[string]*model.DepGraph{}
	_dep := model.NewDepGraphMap(nil, func(s ...string) *model.DepGraph {
		return &model.DepGraph{
			Vendor:  s[0],
			Name:    s[1],
			Version: s[2],
		}
	}).LoadOrStore

	var group, name, version, id string
	f.ReadLine(func(line string) {
		i := strings.Index(line, ":")
		if strings.HasPrefix(line, "#") || i == -1 {
			if id != "" {
				depIdMap[id] = _dep(group, name, version)
			}
			group, name, version, id = "", "", "", ""
			return
		}
		k := strings.TrimSpace(line[:i])
		v := strings.TrimSpace(line[i+1:])
		switch k {
		case "PackageName":
			name = v
		case "PackageVersion":
			version = v
		case "PackageSupplier":
			group = strings.TrimPrefix(v, "Organization: ")
		case "SPDXID":
			id = v
		case "Relationships":
			ids := strings.Split(v, "DEPENDS_ON")
			if len(ids) == 2 {
				parent := strings.TrimSpace(ids[0])
				child := strings.TrimSpace(ids[1])
				depRelation[parent] = append(depRelation[parent], child)
			}
		}
	})

	for parent, children := range depRelation {
		for _, c := range children {
			depIdMap[parent].AppendChild(depIdMap[c])
		}
	}

	root := &model.DepGraph{}
	for _, dep := range depIdMap {
		if len(dep.Parents) == 0 {
			root.AppendChild(dep)
		}
	}

	return root
}

func ParseSpdxJson(f *model.File) *model.DepGraph {
	doc := &model.SpdxDocument{}
	f.OpenReader(func(reader io.Reader) {
		json.NewDecoder(reader).Decode(&doc)
	})
	return parseSpdxDoc(doc)
}

func ParseSpdxXml(f *model.File) *model.DepGraph {
	doc := &model.SpdxDocument{}
	f.OpenReader(func(reader io.Reader) {
		xml.NewDecoder(reader).Decode(&doc)
	})
	return parseSpdxDoc(doc)
}

func parseSpdxDoc(doc *model.SpdxDocument) *model.DepGraph {

	if doc == nil || doc.SPDXVersion == "" {
		return nil
	}

	depIdMap := map[string]*model.DepGraph{}
	_dep := model.NewDepGraphMap(nil, func(s ...string) *model.DepGraph {
		return &model.DepGraph{
			Vendor:  s[0],
			Name:    s[1],
			Version: s[2],
		}
	}).LoadOrStore

	for _, pkg := range doc.Packages {
		depIdMap[pkg.SPDXID] = _dep(pkg.Supplier, pkg.Name, pkg.Version)
	}

	for _, relation := range doc.Relationships {
		depIdMap[relation.SPDXElementID].AppendChild(depIdMap[relation.RelatedSPDXElement])
	}

	return nil
}
