package java

import (
	"context"
	"io"
	"strings"

	"github.com/xmirrorsecurity/opensca-cli/opensca/common"
	"github.com/xmirrorsecurity/opensca-cli/opensca/model"
	"github.com/xmirrorsecurity/opensca-cli/opensca/sca/filter"
)

type Sca struct {
	NotUseMvn    bool
	NotUseStatic bool
}

func (sca Sca) Language() model.Language {
	return model.Lan_Java
}

func (sca Sca) Filter(relpath string) bool {
	return filter.JavaPom(relpath)
}

func (sca Sca) Sca(ctx context.Context, parent *model.File, files []*model.File, call model.ResCallback) {

	// jar包中的pom仅读取pom自身信息 不获取子依赖
	if strings.Contains(parent.Relpath(), ".jar") {
		for _, file := range files {
			if !filter.JavaPom(file.Relpath()) {
				continue
			}
			file.OpenReader(func(reader io.Reader) {
				p := ReadPom(reader)
				if !p.Check() {
					return
				}
				call(file, &model.DepGraph{
					Vendor:  p.GroupId,
					Name:    p.ArtifactId,
					Version: p.Version,
					Path:    file.Relpath(),
				})
			})
		}
	}

	// 记录pom文件
	poms := []*Pom{}
	for _, file := range files {
		if filter.JavaPom(file.Relpath()) {
			file.OpenReader(func(reader io.Reader) {
				pom := ReadPom(reader)
				pom.File = file
				poms = append(poms, pom)
			})
		}
	}

	// 记录不需要静态解析的pom
	var exclusionPom []*Pom

	// 优先尝试调用mvn
	if !sca.NotUseMvn {
		for _, pom := range poms {
			dep := MvnTree(ctx, pom)
			if dep != nil {
				call(pom.File, dep)
				exclusionPom = append(exclusionPom, pom)
			}
		}
	}

	// 静态解析
	if !sca.NotUseStatic {
		ParsePoms(ctx, poms, exclusionPom, func(pom *Pom, root *model.DepGraph) {
			call(pom.File, root)
		})
	}
}

var defaultMavenRepo = []common.RepoConfig{
	{Url: "https://maven.aliyun.com/repository/public"},
	{Url: "https://repo1.maven.org/maven2"},
}

func RegisterMavenRepo(repos ...common.RepoConfig) {
	newRepo := common.TrimRepo(repos...)
	if len(newRepo) > 0 {
		defaultMavenRepo = newRepo
	}
}
