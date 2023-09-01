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

		// 使用parent属性
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

		// exclusion排除子依赖
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

		// dependencyManagement传递scope
		{"3", tool.Dep("", "", "",
			tool.Dep("com.foo", "demo", "1.0"),
			tool.Dep("com.foo", "mod", "1.0",
				tool.Dep("org.springframework", "spring-context", "4.3.6.RELEASE",
					tool.Dep("org.springframework", "spring-aop", "4.3.7.RELEASE"),
					tool.Dep("org.springframework", "spring-beans", "4.3.7.RELEASE"),
					tool.Dep("org.springframework", "spring-core", "4.3.7.RELEASE"),
					tool.DevDep("org.springframework", "spring-expression", "4.3.5.RELEASE"),
				),
			),
		)},

		// 继承parent依赖项 优先使用根pom的属性
		{"4", tool.Dep("", "", "",
			tool.Dep("com.foo", "demo", "1.0",
				tool.Dep("org.springframework", "spring-expression", "4.3.7.RELEASE"),
				tool.Dep("org.springframework", "spring-context", "4.3.7.RELEASE",
					tool.Dep("org.springframework", "spring-aop", "4.3.7.RELEASE"),
					tool.Dep("org.springframework", "spring-beans", "4.3.7.RELEASE"),
					tool.Dep("org.springframework", "spring-core", "4.3.7.RELEASE"),
				),
			),
			tool.Dep("com.foo", "mod", "1.0",
				tool.Dep("org.springframework", "spring-expression", "4.3.5.RELEASE"),
				tool.Dep("org.springframework", "spring-context", "4.3.7.RELEASE",
					tool.Dep("org.springframework", "spring-aop", "4.3.7.RELEASE"),
					tool.Dep("org.springframework", "spring-beans", "4.3.7.RELEASE"),
					tool.Dep("org.springframework", "spring-core", "4.3.7.RELEASE"),
				),
			),
		)},

		// 属性多级引用
		{"5", tool.Dep("", "", "",
			tool.Dep("my.foo", "demo", "1.4.10",
				tool.Dep("org.jetbrains.kotlin", "kotlin-stdlib", "1.4.10",
					tool.Dep("org.jetbrains.kotlin", "kotlin-stdlib-common", "1.4.10"),
					tool.Dep("org.jetbrains", "annotations", "13.0"),
				),
			),
		)},

		// import使用自身pom而非根pom中的属性
		{"6", tool.Dep("", "", "",
			tool.Dep("my.foo", "demo", "1.4.10",
				tool.Dep("org.jetbrains.kotlin", "kotlin-stdlib", "1.6.21",
					tool.Dep("org.jetbrains.kotlin", "kotlin-stdlib-common", "1.6.21"),
					tool.Dep("org.jetbrains", "annotations", "13.0"),
				),
			),
			tool.Dep("my.foo", "demo2", "1.4.10",
				tool.Dep("org.jetbrains.kotlin", "kotlin-stdlib", "1.6.20",
					tool.Dep("org.jetbrains.kotlin", "kotlin-stdlib-common", "1.6.20"),
					tool.Dep("org.jetbrains", "annotations", "13.0"),
				),
			),
		)},

		// 同一个pom中存在厂商和组件相同的依赖时使用后声明的依赖
		{"7", tool.Dep("", "", "",
			tool.Dep("my.foo", "demo", "1.4.10",
				tool.Dep("org.jetbrains.kotlin", "kotlin-stdlib", "1.6.20",
					tool.Dep("org.jetbrains.kotlin", "kotlin-stdlib-common", "1.6.20"),
					tool.Dep("org.jetbrains", "annotations", "13.0"),
				),
			),
		)},

		// 子依赖使用本身的pom而非根pom
		{"8", tool.Dep("", "", "",
			tool.Dep("my.foo", "demo", "1.0",
				tool.Dep("org.redisson", "redisson-spring-boot-starter", "3.18.0",
					tool.Dep("org.redisson", "redisson", "3.18.0"),
				),
			),
		)},
	}

	for _, c := range cases {

		t.Run(c.Path, func(t *testing.T) {

			deps, _ := opensca.RunTask(context.Background(), &opensca.TaskArg{
				DataOrigin: c.Path,
				Sca:        []sca.Sca{java.Sca{NotUseMvn: true}},
			})

			result := &model.DepGraph{}
			for _, dep := range deps {
				result.AppendChild(dep)
			}
			result.ForEachNode(func(p, n *model.DepGraph) bool { n.Path = ""; return true })

			if tool.Diff(result, c.Result) {
				t.Errorf("%s\nres:\n%sstd:\n%s", c.Path, result.Tree(false), c.Result.Tree(false))
			}

		})
	}

}
