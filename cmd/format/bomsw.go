package format

import (
	"encoding/json"
	"io"

	"github.com/xmirrorsecurity/opensca-cli/v3/cmd/detail"
	"github.com/xmirrorsecurity/opensca-cli/v3/opensca/model"
)

func BomSWJson(report Report, out string) {
	outWrite(out, func(w io.Writer) error {
		doc := bomSWDoc(report)
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		return encoder.Encode(doc)
	})
}

func bomSWDoc(report Report) *model.BomSWDocument {

	doc := model.NewBomSWDocument(report.TaskInfo.AppName, "opensca-cli")

	report.DepDetailGraph.ForEach(func(n *detail.DepDetailGraph) bool {

		if n.Name == "" {
			return true
		}

		lics := []string{}
		for _, lic := range n.Licenses {
			lics = append(lics, lic.ShortName)
		}
		doc.AppendComponents(func(swc *model.BomSWComponent) {
			swc.ID = n.Purl()
			swc.Name = n.Name
			swc.Version = n.Version
			swc.License = lics
		})

		children := []string{}
		for _, c := range n.Children {
			if c.Name == "" {
				continue
			}
			children = append(children, c.Purl())
		}
		doc.AppendDependencies(n.Purl(), children)

		return true
	})

	return doc
}
