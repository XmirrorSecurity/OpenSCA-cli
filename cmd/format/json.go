package format

import (
	"encoding/json"

	"github.com/xmirrorsecurity/opensca-cli/cmd/detail"
)

// Json 无重复json
func Dep2Json(dep *detail.DepDetailGraph) string {
	data, _ := json.Marshal(dep)
	return string(data)
}
