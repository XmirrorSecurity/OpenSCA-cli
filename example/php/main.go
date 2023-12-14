package main

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/xmirrorsecurity/opensca-cli/v3/opensca/common"
	"github.com/xmirrorsecurity/opensca-cli/v3/opensca/logs"
	"github.com/xmirrorsecurity/opensca-cli/v3/opensca/model"
	"github.com/xmirrorsecurity/opensca-cli/v3/opensca/sca/php"
)

func init() {

	// register composer repository origin
	php.RegisterComposerOrigin(func(name, version string) *php.ComposerPackage {
		var composer *php.ComposerPackage
		common.DownloadUrlFromRepos(fmt.Sprintf("%s.json", name), func(repo common.RepoConfig, r io.Reader) { composer = php.ReadComposerRepoJson(r, name, version) },
			common.RepoConfig{Url: "http://repo.packagist.org/p2"},
		)
		return composer
	})
}

func main() {

	projectDir := "../../test/php/2"

	sca := php.Sca{}

	var files []*model.File
	filepath.WalkDir(projectDir, func(path string, d fs.DirEntry, err error) error {
		if !sca.Filter(path) {
			return nil
		}
		file := model.NewFile(path, strings.TrimPrefix(path, projectDir))
		files = append(files, file)
		return nil
	})

	sca.Sca(context.TODO(), nil, files, func(file *model.File, root ...*model.DepGraph) {
		for _, dep := range root {
			dep.Build(false, sca.Language())
			logs.Infof("file %s:\n%s", file.Relpath(), dep.Tree(false, false))
		}
	})
}
