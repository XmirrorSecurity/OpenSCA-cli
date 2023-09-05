package sbom

import (
	"context"

	"github.com/xmirrorsecurity/opensca-cli/opensca/model"
	"github.com/xmirrorsecurity/opensca-cli/opensca/sca/filter"
)

type Sca struct{}

func (sca Sca) Language() model.Language {
	return model.Lan_None
}

func (sca Sca) Filter(relpath string) bool {
	return filter.SbomJson(relpath) || filter.SbomXml(relpath) || filter.SbomSpdx(relpath) || filter.SbomDsdx(relpath)
}

func (sca Sca) Sca(ctx context.Context, parent *model.File, files []*model.File) []*model.DepGraph {
	var root []*model.DepGraph
	for _, file := range files {
		if filter.SbomSpdx(file.Relpath) {
			root = append(root, ParseSpdx(file))
		}
		if filter.SbomDsdx(file.Relpath) {
			root = append(root, ParseDsdx(file))
		}
		if filter.SbomJson(file.Relpath) {
			root = append(root, ParseSpdxJson(file))
			root = append(root, ParseCdxJson(file))
			root = append(root, ParseDsdxJson(file))
		}
		if filter.SbomXml(file.Relpath) {
			root = append(root, ParseSpdxXml(file))
			root = append(root, ParseCdxXml(file))
			root = append(root, ParseDsdxXml(file))
		}
	}
	return root
}
