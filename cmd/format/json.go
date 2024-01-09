package format

import (
	"encoding/json"
	"io"
)

func Json(report Report, out string) {
	outWrite(out, func(w io.Writer) error {
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		return encoder.Encode(report)
	})
}
