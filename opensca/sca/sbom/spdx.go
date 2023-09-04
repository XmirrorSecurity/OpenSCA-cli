package sbom

import (
	"encoding/json"
	"encoding/xml"
	"io"

	"github.com/xmirrorsecurity/opensca-cli/opensca/model"
)

func ParseSpdx(f *model.File) *model.DepGraph {
	// TODO
	return nil
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
