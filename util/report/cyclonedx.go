package report

import (
	"io"
	"strings"
	"util/model"

	"github.com/CycloneDX/cyclonedx-go"
)

func buildCycBom(dep *model.DepTree, taskInfo TaskInfo) *cyclonedx.BOM {
	metadata := cyclonedx.Metadata{}
	components := []cyclonedx.Component{}
	dependencies := []cyclonedx.Dependency{}
	q := []*model.DepTree{dep}
	for len(q) > 0 {
		n := q[0]
		q = append(q[1:], n.Children...)
		if n == dep {
			metadata.Component = &cyclonedx.Component{
				BOMRef:     n.Purl(),
				Type:       cyclonedx.ComponentTypeApplication,
				Name:       n.Name,
				Version:    n.VersionStr,
				PackageURL: n.Purl(),
			}
			continue
		}
		if n.Name != "" {
			components = append(components, cyclonedx.Component{
				BOMRef:     n.Purl(),
				Type:       cyclonedx.ComponentTypeLibrary,
				Author:     n.Vendor,
				Name:       n.Name[strings.LastIndex(n.Name, "/")+1:],
				Version:    n.VersionStr,
				PackageURL: n.Purl(),
			})
			var deps []string
			for _, child := range n.Children {
				deps = append(deps, child.Purl())
			}
			dependencies = append(dependencies, cyclonedx.Dependency{
				Ref:          n.Purl(),
				Dependencies: &deps,
			})
		}
	}
	bom := cyclonedx.NewBOM()
	bom.Metadata = &metadata
	bom.Components = &components
	bom.Dependencies = &dependencies
	return bom
}

func CycloneDXJson(writer io.Writer, dep *model.DepTree, taskInfo TaskInfo) {
	bom := buildCycBom(dep, taskInfo)
	cyclonedx.NewBOMEncoder(writer, cyclonedx.BOMFileFormatJSON).SetPretty(true).Encode(bom)
}

func CycloneDXXml(writer io.Writer, dep *model.DepTree, taskInfo TaskInfo) {
	bom := buildCycBom(dep, taskInfo)
	cyclonedx.NewBOMEncoder(writer, cyclonedx.BOMFileFormatXML).SetPretty(true).Encode(bom)
}
