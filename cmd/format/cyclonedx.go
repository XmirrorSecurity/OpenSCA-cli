package format

import (
	"io"
	"strings"

	"github.com/CycloneDX/cyclonedx-go"
	"github.com/xmirrorsecurity/opensca-cli/v3/cmd/detail"
)

func cyclonedxbom(dep *detail.DepDetailGraph) *cyclonedx.BOM {

	metadata := cyclonedx.Metadata{}
	components := []cyclonedx.Component{}
	dependencies := []cyclonedx.Dependency{}

	dep.ForEach(func(n *detail.DepDetailGraph) bool {

		if n == dep {
			metadata.Component = &cyclonedx.Component{
				BOMRef:     n.Purl(),
				Type:       cyclonedx.ComponentTypeApplication,
				Name:       n.Name,
				Version:    n.Version,
				PackageURL: n.Purl(),
			}
			return true
		}

		if n.Name != "" {
			components = append(components, cyclonedx.Component{
				BOMRef:     "ref-" + n.ID,
				Type:       cyclonedx.ComponentTypeLibrary,
				Author:     n.Vendor,
				Name:       n.Name[strings.LastIndex(n.Name, "/")+1:],
				Version:    n.Version,
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

		return true
	})

	bom := cyclonedx.NewBOM()
	bom.Metadata = &metadata
	bom.Components = &components
	bom.Dependencies = &dependencies
	return bom
}

func CycloneDXJson(report Report, out string) {
	bom := cyclonedxbom(report.DepDetailGraph)
	outWrite(out, func(w io.Writer) error {
		return cyclonedx.NewBOMEncoder(w, cyclonedx.BOMFileFormatJSON).SetPretty(true).Encode(bom)
	})
}

func CycloneDXXml(report Report, out string) {
	bom := cyclonedxbom(report.DepDetailGraph)
	outWrite(out, func(w io.Writer) error {
		return cyclonedx.NewBOMEncoder(w, cyclonedx.BOMFileFormatXML).SetPretty(true).Encode(bom)
	})
}
