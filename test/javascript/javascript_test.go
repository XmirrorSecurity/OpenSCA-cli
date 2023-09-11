package javascript

import (
	"testing"

	"github.com/xmirrorsecurity/opensca-cli/opensca/sca/javascript"
	"github.com/xmirrorsecurity/opensca-cli/test/tool"
)

func Test_JavaScript(t *testing.T) {

	std := tool.Dep("", "", "",
		tool.Dep("", "js-test", "1.0.1",
			tool.Dep("", "cliui", "6.0.0",
				tool.Dep("", "string-width", "4.2.3",
					tool.Dep("", "emoji-regex", "8.0.0"),
					tool.Dep("", "is-fullwidth-code-point", "3.0.0"),
				),
				tool.Dep("", "strip-ansi", "6.0.1",
					tool.Dep("", "ansi-regex", "5.0.1"),
				),
				tool.Dep("", "wrap-ansi", "6.2.0",
					tool.Dep("", "ansi-styles", "4.3.0",
						tool.Dep("", "color-convert", "2.0.1",
							tool.Dep("", "color-name", "1.1.4"),
						),
					),
				),
			),
		),
	)

	tool.RunTaskCase(t, javascript.Sca{})([]tool.TaskCase{
		// package.lock
		{Path: "1", Result: std},
		// package.lock v3
		{Path: "2", Result: std},
		// yarn.lock
		{Path: "3", Result: std},
		// node_modules
		{Path: "4", Result: std},
		// simple
		{Path: "5", Result: std},
	})
}
