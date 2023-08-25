package format

import (
	"fmt"
	"io"

	"github.com/xmirrorsecurity/opensca-cli/cmd/detail"
)

func Csv(report Report, out string) {

	table := "Name, Version, Vendor, License, Langauge, PURL\n"

	report.DepDetailGraph.ForEach(func(n *detail.DepDetailGraph) bool {

		licenseTxt := ""
		if len(n.Licenses) > 0 {
			licenseTxt = n.Licenses[0].ShortName
		}

		if n.Name != "" {
			table = table + fmt.Sprintf("%s,%s,%s,%s,%s,%s\n", n.Name, n.Version, n.Vendor, licenseTxt, n.Language, n.Purl())
		}

		return true
	})

	outWrite(out, func(w io.Writer) {
		w.Write([]byte(table))
	})

}
