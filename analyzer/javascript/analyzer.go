/*
 * @Description: javascript解析器
 * @Date: 2021-11-25 19:59:35
 */

package javascript

import (
	"path"
	"util/enum/language"
	"util/filter"
	"util/model"
)

type Analyzer struct{}

// New 创建javascript解析器
func New() Analyzer {
	return Analyzer{}
}

// GetLanguage Get language of Analyzer
func (a Analyzer) GetLanguage() language.Type {
	return language.JavaScript
}

// CheckFile Check if it is a parsable file
func (a Analyzer) CheckFile(filename string) bool {
	return filter.JavaScriptPackageLock(filename) ||
		filter.JavaScriptPackage(filename) ||
		filter.JavaScriptYarnLock(filename)
}

// ParseFile Parse the file
func (a Analyzer) ParseFiles(files []*model.FileInfo) []*model.DepTree {
	deps := []*model.DepTree{}
	pkgMap := map[string]*model.FileInfo{}
	lockMap := map[string]*model.FileInfo{}
	for _, f := range files {
		if filter.JavaScriptPackage(f.Name) {
			pkgMap[path.Dir(f.Name)] = f
		} else if filter.JavaScriptPackageLock(f.Name) {
			lockMap[path.Dir(f.Name)] = f
		}
	}
	for _, f := range files {
		dep := model.NewDepTree(nil)
		dep.Path = f.Name
		if filter.JavaScriptPackage(f.Name) {
			// 检测同目录下是否有package-lock.json文件
			if _, ok := lockMap[path.Dir(f.Name)]; !ok {
				parsePackage(dep, f, true)
			}
		} else if filter.JavaScriptPackageLock(f.Name) {
			// 检测同目录下是否有package.json文件
			if pkg, ok := pkgMap[path.Dir(f.Name)]; !ok {
				parsePackageLock(dep, f, nil)
			} else {
				parsePackageLock(dep, f, parsePackage(dep, pkg, false))
			}
		} else if filter.JavaScriptYarnLock(f.Name) {
			parseYarnLock(dep, f)
		}
		deps = append(deps, dep)
	}
	return deps
}
