/*
 * @Descripation: 文件过滤器
 * @Date: 2021-11-03 15:34:15
 */
package filter

import (
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
)

// erlang
var (
	ErlangRebarLock = filterFunc(strings.HasSuffix, "rebar.lock")
)

// groovy
var (
	GroovyFile = filterFunc(strings.HasSuffix, ".groovy")
)
