package sbom

import (
	"encoding/json"
	"encoding/xml"
	"io"

	"github.com/xmirrorsecurity/opensca-cli/opensca/model"
)

func ParseDsdx(f *model.File) *model.DepGraph {
	// TODO
	return nil
}

func ParseDsdxJson(f *model.File) *model.DepGraph {
	doc := &model.DsdxDocument{}
	f.OpenReader(func(reader io.Reader) {
		json.NewDecoder(reader).Decode(doc)
	})
	return parseDsdxDoc(doc)
}

func ParseDsdxXml(f *model.File) *model.DepGraph {
	doc := &model.DsdxDocument{}
	f.OpenReader(func(reader io.Reader) {
		xml.NewDecoder(reader).Decode(doc)
	})
	return parseDsdxDoc(doc)
}

func parseDsdxDoc(doc *model.DsdxDocument) *model.DepGraph {

	if doc == nil || doc.DSDXVersion == "" {
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

	for _, c := range doc.Components {
		dep := _dep(c.Group, c.Name, c.Version)
		dep.Language = model.Language(c.Language)
		dep.Licenses = c.License
		dep.Path = c.URL
		depIdMap[c.ID] = dep
	}

	for parentId, childrenIds := range doc.Dependencies {
		parent, ok := depIdMap[parentId]
		if !ok {
			continue
		}
		for _, id := range childrenIds {
			parent.AppendChild(depIdMap[id])
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
