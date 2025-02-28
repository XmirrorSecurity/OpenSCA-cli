package sbom

import (
	"context"

	"github.com/xmirrorsecurity/opensca-cli/v3/opensca/model"
	"github.com/xmirrorsecurity/opensca-cli/v3/opensca/sca/filter"
)

type Sca struct{}

func (sca Sca) Language() model.Language {
	return model.Lan_None
}

func (sca Sca) Filter(relpath string) bool {
	return filter.SbomJson(relpath) || filter.SbomXml(relpath) || filter.SbomSpdx(relpath) || filter.SbomDsdx(relpath)
}

func (sca Sca) Sca(ctx context.Context, parent *model.File, files []*model.File, call model.ResCallback) {
	for _, file := range files {
		if filter.SbomSpdx(file.Relpath()) {
			call(file, ParseSpdx(file))
		}
		if filter.SbomDsdx(file.Relpath()) {
			call(file, ParseDsdx(file))
		}
		if filter.SbomDbSbom(file.Relpath()) {
			call(file, ParseDpSbomJson(file))
		}
		if filter.SbomJson(file.Relpath()) {
			call(file, ParseSpdxJson(file))
			call(file, ParseCdxJson(file))
			call(file, ParseDsdxJson(file))
			call(file, ParseDpSbomJson(file))
		}
		if filter.SbomXml(file.Relpath()) {
			call(file, ParseSpdxXml(file))
			call(file, ParseCdxXml(file))
			call(file, ParseDsdxXml(file))
		}
	}
}
