package php

import (
	"testing"

	"github.com/xmirrorsecurity/opensca-cli/v3/opensca/sca/php"
	"github.com/xmirrorsecurity/opensca-cli/v3/test/tool"
)

func Test_Php(t *testing.T) {

	std := tool.Dep("", "",
		tool.Dep("opensca/test", "",
			tool.Dep("http-interop/http-factory-guzzle", "1.2.0",
				tool.Dep("guzzlehttp/psr7", "2.6.1",
					tool.Dep("ralouphie/getallheaders", "3.0.3"),
				),
				tool.Dep("psr/http-factory", "1.0.2"),
			),
			tool.Dep("psr/http-message", "2.0"),
		),
	)

	tool.RunTaskCase(t, php.Sca{})([]tool.TaskCase{

		// composer.lock
		{Path: "1", Result: std},

		// composer.json
		{Path: "2", Result: std},
	})

}
