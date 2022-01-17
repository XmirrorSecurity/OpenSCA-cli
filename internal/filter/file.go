/*
 * @Descripation: 文件过滤器
 * @Date: 2021-11-03 15:34:15
 */
package filter

import (
	"strings"
)

/**
 * @description: 文件名过滤函数
 * @param {func(string,string)bool} strFunc 字符串函数，需要两个字符串参数，返回值为bool类型
 * @param {[]string} args 用作strFunc第二个参数的参数列表
 * @return {func(string) bool} 文件名过滤函数，一个字符串参数，返回bool值。
 */
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
	JavaPom           = filterFunc(strings.HasSuffix, "pom.xml", ".pom")
	JavaPomProperties = filterFunc(strings.HasSuffix, "pom.properties")
)

// javascript相关
var (
	JavaScriptPackageLock = filterFunc(strings.HasSuffix, "package-lock.json")
	JavaScriptPackage     = filterFunc(strings.HasSuffix, "package.json")
)

// php相关
var (
	PhpComposerLock = filterFunc(strings.HasSuffix, "composer.lock")
)

// ruby相关
var (
	RubyGemfileLock = filterFunc(strings.HasSuffix, "Gemfile.lock", "gems.locked")
)
