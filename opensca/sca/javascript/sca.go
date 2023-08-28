package javascript

import (
	"context"
	"io"
	"path"
	"strings"

	"github.com/xmirrorsecurity/opensca-cli/opensca/model"
	"github.com/xmirrorsecurity/opensca-cli/opensca/sca/filter"
)

type Sca struct{}

func (sca Sca) Language() model.Language {
	return model.Lan_JavaScript
}

func (sca Sca) Filter(relpath string) bool {
	return filter.JavaScriptPackageJson(relpath) ||
		filter.JavaScriptPackageLock(relpath) ||
		filter.JavaScriptYarnLock(relpath)
}

func (sca Sca) Sca(ctx context.Context, parent *model.File, files []*model.File) []*model.DepGraph {
	deps := ParsePackageJson(files)
	return deps
}

// files: package.json package-lock.json yarn.lock
func ParsePackageJson(files []*model.File) []*model.DepGraph {

	var root []*model.DepGraph

	// map[name]
	jsonMap := map[string]*PackageJson{}
	// map[name]
	lockMap := map[string]*PackageLock{}
	// map[dirpath]
	nodeMap := map[string]*PackageJson{}
	// map[dirpath]
	yarnMap := map[string]map[string]*YarnLock{}

	path2dir := func(relpath string) string { return path.Dir(relpath) }

	// 将npm相关文件按上述方案分类
	for _, f := range files {
		if filter.JavaScriptYarnLock(f.Relpath) {
			yarnMap[path2dir(f.Relpath)] = ParseYarnLock(f)
		}
		if filter.JavaScriptPackageJson(f.Relpath) {
			var js *PackageJson
			f.OpenReader(func(reader io.Reader) {
				js = readJson[PackageJson](reader)
				if js == nil {
					return
				}
				js.File = f
				if strings.Contains(f.Relpath, "node_modules") {
					nodeMap[path2dir(f.Relpath)] = js
				} else {
					jsonMap[js.Name] = js
				}
			})
		}
		if filter.JavaScriptPackageLock(f.Relpath) {
			f.OpenReader(func(reader io.Reader) {
				lock := readJson[PackageLock](reader)
				if lock == nil {
					return
				}
				lockMap[lock.Name] = lock
			})
		}
	}

	// 遍历非node_modules下的package.json
	for name, js := range jsonMap {

		// 尝试从package-lock.json获取
		if lock, ok := lockMap[name]; ok {
			root = append(root, ParsePackageJsonWithLock(js, lock))
			continue
		}

		// 尝试从yarn.lock获取
		if js.File != nil {
			if yarn, ok := yarnMap[path2dir(js.File.Relpath)]; ok {
				root = append(root, ParsePackageJsonWithYarnLock(js, yarn))
				continue
			}
		}

		// 尝试从node_modules及外部源获取
		root = append(root, ParsePackageJsonWithNode(js, nodeMap))
	}

	return root
}
