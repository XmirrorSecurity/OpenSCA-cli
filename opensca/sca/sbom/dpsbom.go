package sbom

import (
	"encoding/json"
	"io"

	"github.com/xmirrorsecurity/opensca-cli/v3/opensca/model"
)

func ParseDpSbomJson(f *model.File) *model.DepGraph {
	doc := &model.DpSbomDocument{}
	f.OpenReader(func(reader io.Reader) {
		json.NewDecoder(reader).Decode(doc)
	})
	return parseDpSbomDoc(f, doc)
}

func parseDpSbomDoc(f *model.File, doc *model.DpSbomDocument) *model.DepGraph {

	if doc == nil {
		return nil
	}

	depIdMap := map[string]*model.DepGraph{}
	_dep := model.NewDepGraphMap(func(s ...string) string {
		return s[0]
	}, func(s ...string) *model.DepGraph {
		vendor, name, version, language := model.ParsePurl(s[0])
		return &model.DepGraph{
			Vendor:   vendor,
			Name:     name,
			Version:  version,
			Language: language,
		}
	}).LoadOrStore

	for _, pkg := range doc.Packages {
		dep := _dep(pkg.Identifier.Purl)
		dep.Licenses = pkg.License
		depIdMap[pkg.Identifier.Purl] = dep
	}

	for _, dependOn := range doc.Dependencies {
		parent, ok := depIdMap[dependOn.Ref]
		if !ok {
			continue
		}
		for _, dep := range dependOn.DependsOn {
			child, ok := depIdMap[dep.Target]
			if !ok {
				continue
			}
			parent.AppendChild(child)
		}
	}

	root := &model.DepGraph{Path: f.Relpath()}
	for _, dep := range depIdMap {
		if len(dep.Parents) == 0 {
			root.AppendChild(dep)
		}
	}

	return root
}
