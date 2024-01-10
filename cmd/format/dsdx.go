package format

import (
	"encoding/json"
	"encoding/xml"
	"io"

	"github.com/xmirrorsecurity/opensca-cli/v3/cmd/detail"
	"github.com/xmirrorsecurity/opensca-cli/v3/opensca/model"
)

func Dsdx(report Report, out string) {
	outWrite(out, func(w io.Writer) error {
		return dsdxDoc(report).WriteDsdx(w)
	})
}

func DsdxJson(report Report, out string) {
	outWrite(out, func(w io.Writer) error {
		return json.NewEncoder(w).Encode(dsdxDoc(report))
	})
}

func DsdxXml(report Report, out string) {
	outWrite(out, func(w io.Writer) error {
		return xml.NewEncoder(w).Encode(dsdxDoc(report))
	})
}

func dsdxDoc(report Report) *model.DsdxDocument {

	doc := model.NewDsdxDocument(report.TaskInfo.AppName, "opensca-cli")

	report.DepDetailGraph.ForEach(func(n *detail.DepDetailGraph) bool {

		if n.Name == "" {
			return true
		}

		lics := []string{}
		for _, lic := range n.Licenses {
			lics = append(lics, lic.ShortName)
		}
		doc.AppendComponents(n.ID, n.Vendor, n.Name, n.Version, n.Language, lics)

		childrenIds := []string{}
		for _, c := range n.Children {
			if c.Name == "" {
				continue
			}
			childrenIds = append(childrenIds, c.ID)
		}
		doc.AppendDependencies(n.ID, childrenIds)

		return true
	})

	return doc
}
