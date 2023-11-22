package filter

import (
	"path/filepath"
	"strings"
)

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
	JavaPom = filterFunc(strings.HasSuffix, "pom.xml", ".pom")
)

var (
	JavaScriptPackageLock = filterFunc(strings.HasSuffix, "package-lock.json")
	JavaScriptPackageJson = func(filename string) bool {
		return strings.HasSuffix(filename, "package.json")
	}
	JavaScriptYarnLock = filterFunc(strings.HasSuffix, "yarn.lock")
)

var (
	PhpComposer     = filterFunc(strings.HasSuffix, "composer.json")
	PhpComposerLock = filterFunc(strings.HasSuffix, "composer.lock")
)

var (
	RubyGemfileLock = filterFunc(strings.HasSuffix, "Gemfile.lock", "gems.locked")
)

var (
	GoMod     = filterFunc(strings.HasSuffix, "go.mod")
	GoSum     = filterFunc(strings.HasSuffix, "go.sum")
	GoPkgToml = filterFunc(strings.HasSuffix, "Gopkg.toml")
	GoPkgLock = filterFunc(strings.HasSuffix, "Gopkg.lock")
)

var (
	RustCargoLock = filterFunc(strings.HasSuffix, "Cargo.lock")
)

var (
	ErlangRebarLock = filterFunc(strings.HasSuffix, "rebar.lock")
)

var (
	GroovyFile   = filterFunc(strings.HasSuffix, ".groovy")
	GroovyGradle = filterFunc(strings.HasSuffix, ".gradle", ".gradle.kts")
)

var (
	PythonSetup           = filterFunc(strings.HasSuffix, "setup.py")
	PythonPipfile         = filterFunc(strings.HasSuffix, "Pipfile")
	PythonPipfileLock     = filterFunc(strings.HasSuffix, "Pipfile.lock")
	PythonRequirementsTxt = func(filename string) bool {
		return filterFunc(strings.HasSuffix, ".txt")(filename) &&
			filterFunc(strings.Contains, "requirements")(filepath.Base(filename)) && !filterFunc(strings.Contains, "test")(filepath.Base(filename))
	}
	PythonRequirementsIn = filterFunc(strings.HasSuffix, "requirements.in")
)

var (
	SbomSpdx = filterFunc(strings.HasSuffix, ".spdx")
	SbomDsdx = filterFunc(strings.HasSuffix, ".dsdx")
	SbomJson = filterFunc(strings.HasSuffix, ".json")
	SbomXml  = filterFunc(strings.HasSuffix, ".xml")
	// SbomRdf  = filterFunc(strings.HasSuffix, ".rdf")
)

var (
	CompressFile = filterFunc(strings.HasSuffix,
		".zip",
		".jar",
		".rar",
		".tar",
		".gz",
		".bz2",
	)
)
