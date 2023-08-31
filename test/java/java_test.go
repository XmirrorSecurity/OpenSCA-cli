package java

import (
	"context"
	"testing"

	"github.com/xmirrorsecurity/opensca-cli/opensca"
	"github.com/xmirrorsecurity/opensca-cli/opensca/model"
	"github.com/xmirrorsecurity/opensca-cli/opensca/sca"
	"github.com/xmirrorsecurity/opensca-cli/opensca/sca/java"
	"github.com/xmirrorsecurity/opensca-cli/test/tool"
)

func init() {
	java.RegisterRepo(java.MvnRepo{Url: "https://maven.aliyun.com/repository/public"})
}

func Test_Java(t *testing.T) {

	cases := []struct {
		Path   string
		Result *model.DepGraph
	}{

		{"1", tool.Dep("", "", "",
			tool.Dep("com.foo", "demo", "1.0"),
			tool.Dep("com.foo", "mod", "1.0",
				tool.Dep("org.springframework", "spring-context", "4.3.6.RELEASE",
					tool.Dep("org.springframework", "spring-aop", "4.3.6.RELEASE"),
					tool.Dep("org.springframework", "spring-beans", "4.3.6.RELEASE"),
					tool.Dep("org.springframework", "spring-core", "4.3.6.RELEASE"),
					tool.Dep("org.springframework", "spring-expression", "4.3.7.RELEASE"),
				),
			),
		)},

		{"2", tool.Dep("", "", "",
			tool.Dep("com.foo", "demo", "1.0"),
			tool.Dep("com.foo", "mod", "1.0",
				tool.Dep("org.springframework", "spring-context", "4.3.6.RELEASE",
					tool.Dep("org.springframework", "spring-beans", "4.3.6.RELEASE"),
					tool.Dep("org.springframework", "spring-core", "4.3.6.RELEASE"),
					tool.Dep("org.springframework", "spring-expression", "4.3.7.RELEASE"),
				),
			),
		)},
	}

	for _, c := range cases {

		t.Run(c.Path, func(t *testing.T) {

			deps, _ := opensca.RunTask(context.Background(), &opensca.TaskArg{
				DataOrigin: c.Path,
				Name:       "test",
				Sca:        []sca.Sca{java.Sca{NotUseMvn: true}},
			})

			result := &model.DepGraph{}
			for _, dep := range deps {
				result.AppendChild(dep)
			}

			if tool.Diff(result, c.Result) {
				t.Errorf("%s\nres:\n%sstd:\n%s", c.Path, result.Tree(false), c.Result.Tree(false))
			}

		})
	}

}
