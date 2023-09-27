package java

import (
	"testing"

	"github.com/xmirrorsecurity/opensca-cli/opensca/common"
	"github.com/xmirrorsecurity/opensca-cli/opensca/sca/java"
	"github.com/xmirrorsecurity/opensca-cli/test/tool"
)

func Test_Java(t *testing.T) {
	java.RegisterMavenRepo(common.RepoConfig{Url: "https://maven.aliyun.com/repository/public"})
	tool.RunTaskCase(t, java.Sca{NotUseMvn: true})([]tool.TaskCase{

		// 使用parent属性
		{Path: "1", Result: tool.Dep("", "",
			tool.Dep3("com.foo", "demo", "1.0"),
			tool.Dep3("com.foo", "mod", "1.0",
				tool.Dep3("org.springframework", "spring-context", "4.3.6.RELEASE",
					tool.Dep3("org.springframework", "spring-aop", "4.3.6.RELEASE"),
					tool.Dep3("org.springframework", "spring-beans", "4.3.6.RELEASE"),
					tool.Dep3("org.springframework", "spring-core", "4.3.6.RELEASE"),
					tool.Dep3("org.springframework", "spring-expression", "4.3.7.RELEASE"),
				),
			),
		)},

		// exclusion排除子依赖
		{Path: "2", Result: tool.Dep("", "",
			tool.Dep3("com.foo", "demo", "1.0"),
			tool.Dep3("com.foo", "mod", "1.0",
				tool.Dep3("org.springframework", "spring-context", "4.3.6.RELEASE",
					tool.Dep3("org.springframework", "spring-beans", "4.3.6.RELEASE"),
					tool.Dep3("org.springframework", "spring-core", "4.3.6.RELEASE"),
					tool.Dep3("org.springframework", "spring-expression", "4.3.7.RELEASE"),
				),
			),
		)},

		// dependencyManagement传递scope
		{Path: "3", Result: tool.Dep("", "",
			tool.Dep3("com.foo", "demo", "1.0"),
			tool.Dep3("com.foo", "mod", "1.0",
				tool.Dep3("org.springframework", "spring-context", "4.3.6.RELEASE",
					tool.Dep3("org.springframework", "spring-aop", "4.3.7.RELEASE"),
					tool.Dep3("org.springframework", "spring-beans", "4.3.7.RELEASE"),
					tool.Dep3("org.springframework", "spring-core", "4.3.7.RELEASE"),
					tool.DevDep3("org.springframework", "spring-expression", "4.3.5.RELEASE"),
				),
			),
		)},

		// 继承parent依赖项 优先使用根pom的属性
		{Path: "4", Result: tool.Dep("", "",
			tool.Dep3("com.foo", "demo", "1.0",
				tool.Dep3("org.springframework", "spring-expression", "4.3.7.RELEASE"),
				tool.Dep3("org.springframework", "spring-context", "4.3.7.RELEASE",
					tool.Dep3("org.springframework", "spring-aop", "4.3.7.RELEASE"),
					tool.Dep3("org.springframework", "spring-beans", "4.3.7.RELEASE"),
					tool.Dep3("org.springframework", "spring-core", "4.3.7.RELEASE"),
				),
			),
			tool.Dep3("com.foo", "mod", "1.0",
				tool.Dep3("org.springframework", "spring-expression", "4.3.5.RELEASE"),
				tool.Dep3("org.springframework", "spring-context", "4.3.7.RELEASE",
					tool.Dep3("org.springframework", "spring-aop", "4.3.7.RELEASE"),
					tool.Dep3("org.springframework", "spring-beans", "4.3.7.RELEASE"),
					tool.Dep3("org.springframework", "spring-core", "4.3.7.RELEASE"),
				),
			),
		)},

		// 属性多级引用
		{Path: "5", Result: tool.Dep("", "",
			tool.Dep3("my.foo", "demo", "1.4.10",
				tool.Dep3("org.jetbrains.kotlin", "kotlin-stdlib", "1.4.10",
					tool.Dep3("org.jetbrains.kotlin", "kotlin-stdlib-common", "1.4.10"),
					tool.Dep3("org.jetbrains", "annotations", "13.0"),
				),
			),
		)},

		// import使用自身pom而非根pom中的属性
		{Path: "6", Result: tool.Dep("", "",
			tool.Dep3("my.foo", "demo", "1.4.10",
				tool.Dep3("org.jetbrains.kotlin", "kotlin-stdlib", "1.6.21",
					tool.Dep3("org.jetbrains.kotlin", "kotlin-stdlib-common", "1.6.20"),
					tool.Dep3("org.jetbrains", "annotations", "13.0"),
				),
			),
		)},

		// 同一个pom中存在厂商和组件相同的依赖时使用后声明的依赖
		{Path: "7", Result: tool.Dep("", "",
			tool.Dep3("my.foo", "demo", "1.4.10",
				tool.Dep3("org.jetbrains.kotlin", "kotlin-stdlib", "1.6.20",
					tool.Dep3("org.jetbrains.kotlin", "kotlin-stdlib-common", "1.6.20"),
					tool.Dep3("org.jetbrains", "annotations", "13.0"),
				),
			),
		)},

		// 子依赖使用本身的pom而非根pom
		{Path: "8", Result: tool.Dep("", "",
			tool.Dep3("my.foo", "demo", "1.0",
				tool.Dep3("org.redisson", "redisson-spring-boot-starter", "3.18.0",
					tool.Dep3("org.redisson", "redisson", "3.18.0"),
				),
			),
		)},

		// 存在多个厂商和组件名相同的间接依赖时保留最新声明的
		{Path: "9", Result: tool.Dep("", "",
			tool.Dep3("my.foo", "demo", "1.0",
				tool.Dep3("org.springframework.boot", "spring-boot-starter-json", "2.7.14",
					tool.Dep3("com.fasterxml.jackson.core", "jackson-databind", "2.13.5",
						tool.Dep3("com.fasterxml.jackson.core", "jackson-annotations", "2.13.5"),
					),
				),
				tool.Dep3("com.alibaba.nacos", "nacos-client", "2.0.4",
					tool.Dep3("com.fasterxml.jackson.core", "jackson-core", "2.12.2"),
				),
			),
		)},

		// 项目中pom属性多层传递
		{Path: "10", Result: tool.Dep("", "",
			tool.Dep3("com.foo", "demo", "1.0"),
			tool.Dep3("com.foo", "mod", "2.0"),
			tool.Dep3("com.foo", "mod2", "2.0"),
		)},
	})
}
