package python

import (
	"testing"

	"github.com/xmirrorsecurity/opensca-cli/v3/opensca/sca/python"
	"github.com/xmirrorsecurity/opensca-cli/v3/test/tool"
)

func Test_Python(t *testing.T) {

	tool.RunTaskCase(t, python.Sca{})([]tool.TaskCase{

		// rquirements.txt
		{Path: "1", Result: tool.Dep("", "", tool.Dep("", "",
			tool.Dep("elasticsearch", "8.9.0",
				tool.Dep("elastic-transport", "8.4.0",
					tool.Dep("certifi", "2023.7.22"),
					tool.Dep("urllib3", "1.26.16"),
				),
			),
		))},

		// Pipfile
		{Path: "2", Result: tool.Dep("", "", tool.Dep("", "",
			tool.Dep("elasticsearch", "8.9.0",
				tool.Dep("elastic-transport", "8.4.0",
					tool.Dep("certifi", "2023.7.22"),
					tool.Dep("urllib3", "1.26.16"),
				),
			),
		))},

		// Pipfile.lock
		{Path: "3", Result: tool.Dep("", "", tool.Dep("", "",
			tool.Dep("elasticsearch", "8.9.0"),
			tool.Dep("elastic-transport", "8.4.0"),
			tool.Dep("certifi", "2023.7.22"),
			tool.Dep("urllib3", "1.26.16"),
		))},
	})

}
