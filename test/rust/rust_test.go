package rust

import (
	"testing"

	"github.com/xmirrorsecurity/opensca-cli/opensca/sca/rust"
	"github.com/xmirrorsecurity/opensca-cli/test/tool"
)

func Test_Rust(t *testing.T) {
	tool.RunTaskCase(t, rust.Sca{})([]tool.TaskCase{

		// Cargo.lock
		{Path: "1", Result: tool.Dep("", "", "",
			tool.Dep("", "foo", "0.1.0",
				tool.Dep("", "tokio", "1.28.0",
					tool.Dep("", "windows-sys", "0.48.0"),
				),
				tool.Dep("", "windows-targets", "0.48.0",
					tool.Dep("", "windows_x86_64_gnu", "0.48.0"),
				),
			),
		)},
	})
}
