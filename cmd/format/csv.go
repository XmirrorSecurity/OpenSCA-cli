package format

import (
	"fmt"
	"io"
	"strings"

	"github.com/xmirrorsecurity/opensca-cli/v3/cmd/detail"
)

func Csv(report Report, out string) {

	table := "Name, Version, Vendor, License, Language, PURL\n"

	report.DepDetailGraph.ForEach(func(n *detail.DepDetailGraph) bool {

		licenseTxt := ""
		if len(n.Licenses) > 0 {
			licenseTxt = n.Licenses[0].ShortName
		}

		formatCsv := func(s string) string {
			if strings.Contains(s, `"`) {
				s = strings.ReplaceAll(s, `"`, `""`)
			}
			if strings.Contains(s, `,`) {
				s = fmt.Sprintf(`"%s"`, s)
			}
			return s
		}

		if n.Name != "" {
			table = table + fmt.Sprintf("%s,%s,%s,%s,%s,%s\n",
				formatCsv(n.Name),
				formatCsv(n.Version),
				formatCsv(n.Vendor),
				formatCsv(licenseTxt),
				formatCsv(n.Language),
				formatCsv(n.Purl()),
			)
		}

		return true
	})

	outWrite(out, func(w io.Writer) error {
		_, err := w.Write([]byte(table))
		return err
	})

}
