package format

import (
	"encoding/json"
	"encoding/xml"
	"io"

	"github.com/xmirrorsecurity/opensca-cli/v3/cmd/detail"
	"github.com/xmirrorsecurity/opensca-cli/v3/opensca/logs"
	"github.com/xmirrorsecurity/opensca-cli/v3/opensca/model"
)

func Spdx(report Report, out string) {
	outWrite(out, func(w io.Writer) {
		err := spdxDoc(report).WriteSpdx(w)
		if err != nil {
			logs.Warn(err)
		}
	})
}

func SpdxJson(report Report, out string) {
	outWrite(out, func(w io.Writer) {
		json.NewEncoder(w).Encode(spdxDoc(report))
	})
}

func SpdxXml(report Report, out string) {
	outWrite(out, func(w io.Writer) {
		xml.NewEncoder(w).Encode(spdxDoc(report))
	})
}

func spdxDoc(report Report) *model.SpdxDocument {

	doc := model.NewSpdxDocument(report.TaskInfo.AppName)

	report.DepDetailGraph.ForEach(func(n *detail.DepDetailGraph) bool {

		if n.Name == "" {
			return true
		}

		lics := []string{}
		for _, lic := range n.Licenses {
			lics = append(lics, lic.ShortName)
		}
		doc.AddPackage(n.ID, n.Vendor, n.Name, n.Version, model.Language(n.Language), lics)

		for _, c := range n.Children {
			if c.Name == "" {
				continue
			}
			doc.AddRelation(n.ID, c.ID)
		}

		return true
	})

	return doc
}
