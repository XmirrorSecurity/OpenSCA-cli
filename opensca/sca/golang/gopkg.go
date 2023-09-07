package golang

import (
	"io"

	"github.com/BurntSushi/toml"
	"github.com/xmirrorsecurity/opensca-cli/opensca/model"
)

type GopkgToml struct {
	Constraint []struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	} `toml:"constraint"`
}

type GopkgLock struct {
	Projects []struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	} `toml:"projects"`
}

func ParseGopkgToml(f *model.File) *model.DepGraph {
	root := &model.DepGraph{Path: f.Path()}
	gopkg := GopkgToml{}
	f.OpenReader(func(reader io.Reader) {
		toml.NewDecoder(reader).Decode(&gopkg)
	})
	for _, dep := range gopkg.Constraint {
		root.AppendChild(&model.DepGraph{Name: dep.Name, Version: dep.Version})
	}
	return root
}

func ParseGopkgLock(f *model.File) *model.DepGraph {
	root := &model.DepGraph{Path: f.Path()}
	pkglock := GopkgLock{}
	f.OpenReader(func(reader io.Reader) {
		toml.NewDecoder(reader).Decode(&pkglock)
	})
	for _, dep := range pkglock.Projects {
		root.AppendChild(&model.DepGraph{Name: dep.Name, Version: dep.Version})
	}
	return root
}
