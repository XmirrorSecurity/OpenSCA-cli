package sbom

import (
	"encoding/json"
	"encoding/xml"
	"io"

	"github.com/CycloneDX/cyclonedx-go"
	"github.com/xmirrorsecurity/opensca-cli/opensca/model"
)

func ParseCdxJson(f *model.File) *model.DepGraph {
	bom := &cyclonedx.BOM{}
	f.OpenReader(func(reader io.Reader) {
		json.NewDecoder(reader).Decode(&bom)
	})
	return parseCdxBom(bom)
}

func ParseCdxXml(f *model.File) *model.DepGraph {
	bom := &cyclonedx.BOM{}
	f.OpenReader(func(reader io.Reader) {
		xml.NewDecoder(reader).Decode(&bom)
	})
	return parseCdxBom(bom)
}

func parseCdxBom(bom *cyclonedx.BOM) *model.DepGraph {

	if bom == nil || bom.BOMFormat != "CycloneDX" {
		return nil
	}

	if bom.Components == nil || len(*bom.Components) == 0 {
		return nil
	}

	depRefMap := map[string]*model.DepGraph{}
	_dep := model.NewDepGraphMap(nil, func(s ...string) *model.DepGraph {
		return &model.DepGraph{
			Vendor:  s[0],
			Name:    s[1],
			Version: s[2],
		}
	}).LoadOrStore

	for _, d := range *bom.Components {
		if d.Name == "" {
			continue
		}
		depRefMap[d.BOMRef] = _dep(d.Author, d.Name, d.Version)
	}

	for _, d := range *bom.Dependencies {
		dep, ok := depRefMap[d.Ref]
		if !ok || d.Dependencies == nil {
			continue
		}
		for _, subRef := range *d.Dependencies {
			dep.AppendChild(depRefMap[subRef])
		}
	}

	root := &model.DepGraph{}
	for _, dep := range depRefMap {
		if len(dep.Parents) == 0 {
			root.AppendChild(dep)
		}
	}

	return root
}
