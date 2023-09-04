package format

import (
	"encoding/json"
	"encoding/xml"
	"io"

	"github.com/xmirrorsecurity/opensca-cli/cmd/detail"
	"github.com/xmirrorsecurity/opensca-cli/opensca/logs"
	"github.com/xmirrorsecurity/opensca-cli/opensca/model"
)

func Dsdx(report Report, out string) {
	outWrite(out, func(w io.Writer) {
		err := dsdxDoc(report).WriteDsdx(w)
		if err != nil {
			logs.Warn(err)
		}
	})
}

func DsdxJson(report Report, out string) {
	outWrite(out, func(w io.Writer) {
		json.NewEncoder(w).Encode(spdxDoc(report))
	})
}

func DsdxXml(report Report, out string) {
	outWrite(out, func(w io.Writer) {
		xml.NewEncoder(w).Encode(spdxDoc(report))
	})
}

func dsdxDoc(report Report) *model.DsdxDocument {

	doc := model.NewDsdxDocument(report.AppName, "opensca-cli")

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

	return nil
}
