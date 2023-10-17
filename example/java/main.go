package main

import (
	"context"
	"io"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/xmirrorsecurity/opensca-cli/opensca/common"
	"github.com/xmirrorsecurity/opensca-cli/opensca/logs"
	"github.com/xmirrorsecurity/opensca-cli/opensca/model"
	"github.com/xmirrorsecurity/opensca-cli/opensca/sca/java"
)

func init() {

	// register default maven repository
	java.RegisterMavenRepo(
		common.RepoConfig{Url: "https://maven.aliyun.com/repository/public"},
		common.RepoConfig{Url: "https://repo1.maven.org/maven2"},
		// custom maven repostiory like nexus
		common.RepoConfig{Url: "", Username: "", Password: ""},
	)

	// register maven component repository origin
	java.RegisterMavenOrigin(func(groupId, artifactId, version string) *java.Pom {
		var pom *java.Pom
		java.DownloadPomFromRepo(java.PomDependency{GroupId: groupId, ArtifactId: artifactId, Version: version}, func(r io.Reader) { pom = java.ReadPom(r) })
		return pom
	})
}

func main() {

	projectDir := "../../test/java/9"

	// find pom files
	var poms []*java.Pom
	filepath.WalkDir(projectDir, func(path string, d fs.DirEntry, err error) error {
		if !strings.HasSuffix(path, "pom.xml") {
			return nil
		}
		pomfile := model.NewFile(path, strings.TrimPrefix(path, projectDir))
		pomfile.OpenReader(func(reader io.Reader) {
			pom := java.ReadPom(reader)
			pom.File = pomfile
			poms = append(poms, pom)
		})
		return nil
	})

	// pure static parse pom
	java.ParsePoms(context.TODO(), poms, nil, func(pom *java.Pom, root *model.DepGraph) {
		root.Build(false, model.Lan_Java)
		logs.Infof("file %s:\n%s", pom.File.Relpath(), root.Tree(false, false))
	})
}
