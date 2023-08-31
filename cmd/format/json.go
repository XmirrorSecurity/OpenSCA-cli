package format

import (
	"encoding/json"
	"io"
)

func Json(report Report, out string) {
	outWrite(out, func(w io.Writer) {
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		encoder.Encode(report)
	})
}
