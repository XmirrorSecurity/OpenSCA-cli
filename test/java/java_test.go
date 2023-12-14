package java

import (
	"testing"

	"github.com/xmirrorsecurity/opensca-cli/v3/opensca/common"
	"github.com/xmirrorsecurity/opensca-cli/v3/opensca/sca/java"
	"github.com/xmirrorsecurity/opensca-cli/v3/test/tool"
)

var cases = []tool.TaskCase{

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

	// 继承parent 优先使用根pom的属性及DependencyManagement
	{Path: "4", Result: tool.Dep("", "",
		tool.Dep3("com.foo", "demo", "1.0",
			tool.Dep3("org.springframework", "spring-expression", "4.3.6.RELEASE"),
		),
		tool.Dep3("com.foo", "mod", "1.0",
			tool.Dep3("org.springframework", "spring-expression", "4.3.4.RELEASE"),
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
		tool.Dep3("com.foo", "mod", "1.0",
			tool.Dep3("com.alibaba.nacos", "nacos-all", "2.0.3"),
		),
		tool.Dep3("com.foo", "demo", "1.0",
			tool.Dep3("com.foo", "mod", "1.0",
				tool.Dep3("com.alibaba.nacos", "nacos-all", "2.0.3"),
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

	// 支持relativePath
	{Path: "11", Result: tool.Dep("", "",
		tool.Dep3("com", "a", "2.0"),
		tool.Dep3("com", "b", "2.0",
			tool.Dep3("com", "xx", "2.0"),
		),
		tool.Dep3("com", "c", "1.0",
			tool.Dep3("com", "xx", "1.0"),
		),
		tool.Dep3("com", "d", "1.0",
			tool.Dep3("com", "xx", "1.0"),
		),
	)},

	// 间接依赖继承自身pom
	{Path: "12", Result: tool.Dep("", "",
		tool.Dep3("my.foo", "demo", "1.0",
			tool.Dep3("com.google.guava", "guava", "22.0",
				tool.Dep3("com.google.errorprone", "error_prone_annotations", "2.0.18"),
			),
		),
	)},

	// 支持profiles
	{Path: "13", Result: tool.Dep("", "",
		tool.Dep3("my.foo", "demo", "1.0",
			tool.Dep3("org.jboss.resteasy", "resteasy-jaxrs", "3.15.6.Final",
				tool.Dep3("commons-io", "commons-io", "2.10.0"),
			),
		),
	)},

	// 子依赖需要先解析变量再尝试用dependencyManagement补全
	{Path: "14", Result: tool.Dep("", "",
		tool.Dep3("my.foo", "demo", "1.0",
			tool.Dep3("org.glassfish.jaxb", "jaxb-runtime", "2.3.3-b02",
				tool.Dep3("org.glassfish.jaxb", "txw2", "2.3.3-b02"),
			),
		),
	)},
}

func Test_JavaWithStatic(t *testing.T) {
	//init http client
	common.InitHttpClient(false)

	java.RegisterMavenRepo(common.RepoConfig{Url: "https://maven.aliyun.com/repository/public"})
	tool.RunTaskCase(t, java.Sca{NotUseMvn: true})(cases)
}

func Test_JavaWithMvn(t *testing.T) {
	tool.RunTaskCase(t, java.Sca{NotUseStatic: true})(cases)
}
