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

	if bom == nil {
		return nil
	}

	if bom.BOMFormat == "" && bom.XMLNS == "" {
		return nil
	}

	if bom.Components == nil || len(*bom.Components) == 0 {
		return nil
	}

	depRefMap := map[string]*model.DepGraph{}
	_dep := model.NewDepGraphMap(func(s ...string) string {
		return s[0]
	}, func(s ...string) *model.DepGraph {
		return &model.DepGraph{
			Vendor:  s[1],
			Name:    s[2],
			Version: s[3],
		}
	}).LoadOrStore

	for _, d := range *bom.Components {

		if d.PackageURL != "" {
			vendor, name, version, language := model.ParsePurl(d.PackageURL)
			if name != "" {
				dep := _dep(d.BOMRef, vendor, name, version)
				dep.Language = language
				depRefMap[d.BOMRef] = dep
				continue
			}
		}

		if d.Name != "" {
			depRefMap[d.BOMRef] = _dep(d.BOMRef, d.Author, d.Name, d.Version)
		}
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
