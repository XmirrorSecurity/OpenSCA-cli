package ruby

import (
	"testing"

	"github.com/xmirrorsecurity/opensca-cli/opensca/sca/ruby"
	"github.com/xmirrorsecurity/opensca-cli/test/tool"
)

func Test_Ruby(t *testing.T) {
	tool.RunTaskCase(t, ruby.Sca{})([]tool.TaskCase{

		// Gemfile.lock
		{Path: "1", Result: tool.Dep("", "", "",
			tool.Dep("", "em-http-request", "1.1.7",
				tool.Dep("", "addressable", "2.8.5",
					tool.Dep("", "public_suffix", "5.0.3"),
				),
				tool.Dep("", "cookiejar", "0.3.3"),
				tool.Dep("", "em-socksify", "0.3.2"),
				tool.Dep("", "eventmachine", "1.2.7"),
				tool.Dep("", "http_parser.rb", "0.8.0"),
			),
		)},
	})
}
