/*
 * @Descripation: 文件过滤器
 * @Date: 2021-11-03 15:34:15
 */
package filter

import (
	"fmt"
	"path"
	"regexp"
	"strings"
)

// filterFunc 文件名过滤函数
func filterFunc(strFunc func(string, string) bool, args ...string) func(string) bool {
	return func(filename string) bool {
		for _, suffix := range args {
			if strFunc(filename, suffix) {
				return true
			}
		}
		return false
	}
}

var (
	Rar    = filterFunc(strings.HasSuffix, ".rar")
	Zip    = filterFunc(strings.HasSuffix, ".zip")
	Tar    = filterFunc(strings.HasSuffix, ".tar")
	TarGz  = filterFunc(strings.HasSuffix, ".tar.gz", ".tgz")
	TarBz2 = filterFunc(strings.HasSuffix, ".tar.bz2")
	Jar    = func(filename string) bool {
		return filterFunc(strings.HasSuffix, ".jar", ".war")(filename) &&
			!filterFunc(strings.HasSuffix, "sources.jar", "javadoc.jar", "src.jar", "doc.jar")(filename) &&
			!strings.HasPrefix(filename, "._")
	}
	AllPkg = func(filename string) bool {
		return Rar(filename) ||
			Zip(filename) ||
			Tar(filename) ||
			TarGz(filename) ||
			TarBz2(filename) ||
			Jar(filename)
	}
)

// java相关
var (
	JavaPom = filterFunc(strings.HasSuffix, "pom.xml", ".pom")
)

// javascript相关
var (
	JavaScriptPackageLock = filterFunc(strings.HasSuffix, "package-lock.json")
	JavaScriptPackage     = filterFunc(strings.HasSuffix, "package.json")
	JavaScriptYarnLock    = filterFunc(strings.HasSuffix, "yarn.lock")
)

// php相关
var (
	PhpComposer     = filterFunc(strings.HasSuffix, "composer.json")
	PhpComposerLock = filterFunc(strings.HasSuffix, "composer.lock")
)

// ruby相关
var (
	RubyGemfile     = filterFunc(strings.HasSuffix, "Gemfile")
	RubyGemfileLock = filterFunc(strings.HasSuffix, "Gemfile.lock", "gems.locked")
)

// golang
var (
	GoMod = filterFunc(strings.HasSuffix, "go.mod")
	GoSum = filterFunc(strings.HasSuffix, "go.sum")
)

// rust
var (
	RustCargoLock = filterFunc(strings.HasSuffix, "Cargo.lock")
	RustCargoToml = filterFunc(strings.HasSuffix, "Cargo.toml")
)

// erlang
var (
	ErlangRebarLock = filterFunc(strings.HasSuffix, "rebar.lock")
)

// groovy
var (
	GroovyFile   = filterFunc(strings.HasSuffix, ".groovy")
	GroovyGradle = filterFunc(strings.HasSuffix, ".gradle", ".gradle.kts")
)

// python
var (
	PythonSetup           = filterFunc(strings.HasSuffix, "setup.py")
	PythonPyproject       = filterFunc(strings.HasSuffix, "pyproject.toml")
	PythonPipfile         = filterFunc(strings.HasSuffix, "Pipfile")
	PythonPipfileLock     = filterFunc(strings.HasSuffix, "Pipfile.lock")
	PythonRequirementsTxt = func(filename string) bool {
		return filterFunc(strings.HasSuffix, ".txt")(filename) &&
			filterFunc(strings.Contains, "requirements")(path.Base(filename)) &&
			!filterFunc(strings.Contains, "test")(path.Base(filename)) &&
			!filterFunc(strings.Contains, "dev")(path.Base(filename))
	}
	PythonRequirementsIn = filterFunc(strings.HasSuffix, "requirements.in")
	// PythonSetupCfg       = filterFunc(strings.HasSuffix, "setup.cfg")
)

// 用于筛选可能有copyright信息的文件
var (
	LicenseFileNames = []string{
		"li[cs]en[cs]e(s?)",
		"legal",
		"copy(left|right|ing)",
		"unlicense",
		"l?gpl([-_ v]?)(\\d\\.?\\d)?",
		"bsd",
		"mit",
		"apache",
	}
	LicenseFileRe = regexp.MustCompile(
		fmt.Sprintf("^(|.*[-_. ])(%s)(|[-_. ].*)$",
			strings.Join(LicenseFileNames, "|")))
)

func CheckLicense(name string) bool {
	return LicenseFileRe.MatchString(strings.ToLower(path.Base(name)))
}
