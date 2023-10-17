package main

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"log"
	"path/filepath"
	"strings"

	"github.com/xmirrorsecurity/opensca-cli/opensca/logs"
	"github.com/xmirrorsecurity/opensca-cli/opensca/model"
	"github.com/xmirrorsecurity/opensca-cli/opensca/sca/java"
)

func main() {

	// maven project dir
	mvnProjectDir := "../../test/java/9"

	// find pom files
	var poms []*java.Pom
	filepath.WalkDir(mvnProjectDir, func(path string, d fs.DirEntry, err error) error {
		if strings.HasSuffix(path, "pom.xml") {
			pomfile := model.NewFile(path, strings.TrimPrefix(path, mvnProjectDir))
			pomfile.OpenReader(func(reader io.Reader) {
				pom := java.ReadPom(reader)
				pom.File = pomfile
				poms = append(poms, pom)
			})
		}
		return nil
	})

	// pure static parse pom
	java.ParsePoms(context.TODO(), poms, nil, func(pom *java.Pom, root *model.DepGraph) {
		root.Build(false, model.Lan_Java)
		logs.Infof("file %s:\n%s", pom.File.Relpath(), root.Tree(false, false))
	})
}

func init() {

	// register log format
	logs.RegisterOut(func(level logs.Level, format string, v ...any) {
		if format == "" {
			log.Output(3, fmt.Sprint(v...))
		} else {
			log.Output(3, fmt.Sprintf(format, v...))
		}
	})

	// register maven component repository origin
	java.RegisterMavenOrigin(func(groupId, artifactId, version string) *java.Pom {
		var pom *java.Pom
		java.DownloadPomFromRepo(java.PomDependency{GroupId: groupId, ArtifactId: artifactId, Version: version}, func(r io.Reader) { pom = java.ReadPom(r) })
		return pom
	})
}
