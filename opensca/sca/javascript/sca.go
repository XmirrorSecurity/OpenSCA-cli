package javascript

import (
	"context"
	"io"
	"path/filepath"
	"strings"

	"github.com/xmirrorsecurity/opensca-cli/opensca/common"
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

func (sca Sca) Sca(ctx context.Context, parent *model.File, files []*model.File, call model.ResCallback) {

	// map[dirpath]
	jsonMap := map[string]*PackageJson{}
	// map[dirpath]
	lockMap := map[string]*PackageLock{}
	// map[dirpath]
	nodeMap := map[string]*PackageJson{}
	// map[dirpath]
	yarnMap := map[string]map[string]*YarnLock{}

	// 将npm相关文件按上述方案分类
	for _, f := range files {

		dir := filepath.Dir(strings.ReplaceAll(f.Relpath(), `\`, `/`))

		if filter.JavaScriptYarnLock(f.Relpath()) {
			yarnMap[dir] = ParseYarnLock(f)
		}

		if filter.JavaScriptPackageJson(f.Relpath()) {
			var js *PackageJson
			f.OpenReader(func(reader io.Reader) {
				js = readJson[PackageJson](reader)
				if js == nil {
					return
				}
				// 记录 resolutions
				for k, v := range js.Resolutions {
					k = k[strings.LastIndex(k, "/")+1:]
					js.Dependencies[k] = v
				}
				js.File = f
				if strings.Contains(f.Relpath(), "node_modules") {
					nodeMap[dir] = js
				} else {
					jsonMap[dir] = js
				}
			})
		}

		if filter.JavaScriptPackageLock(f.Relpath()) {
			f.OpenReader(func(reader io.Reader) {
				lock := readJson[PackageLock](reader)
				if lock == nil {
					return
				}
				lockMap[dir] = lock
			})
		}
	}

	// 遍历非node_modules下的package.json
	for dir, js := range jsonMap {

		// 尝试从package-lock.json获取
		if lock, ok := lockMap[dir]; ok {
			call(js.File, ParsePackageJsonWithLock(js, lock))
			continue
		}

		// 尝试从yarn.lock获取
		if js.File != nil {
			if yarn, ok := yarnMap[dir]; ok {
				call(js.File, ParsePackageJsonWithYarnLock(js, yarn))
				continue
			}
		}

		select {
		case <-ctx.Done():
			return
		default:
		}

		// 尝试从node_modules及外部源获取
		call(js.File, ParsePackageJsonWithNode(js, nodeMap))
	}
}

var defaultNpmRepo = []common.RepoConfig{
	{Url: "https://r.cnpmjs.org/"},
}

func RegisterNpmRepo(repos ...common.RepoConfig) {
	newRepo := common.TrimRepo(repos...)
	if len(newRepo) > 0 {
		defaultNpmRepo = newRepo
	}
}
