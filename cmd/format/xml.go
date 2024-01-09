package format

import (
	"encoding/xml"
	"io"
)

func Xml(report Report, out string) {
	outWrite(out, func(w io.Writer) error {
		return xml.NewEncoder(w).Encode(report)
	})
}
