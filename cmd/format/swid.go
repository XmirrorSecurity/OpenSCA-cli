package format

import (
	"archive/zip"
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
	"path/filepath"
	"strings"

	"github.com/xmirrorsecurity/opensca-cli/v3/cmd/detail"
	"github.com/xmirrorsecurity/opensca-cli/v3/opensca/logs"

	"github.com/veraison/swid"
)

func swidZip(out string, report Report, writeFunc func(tag *swid.SoftwareIdentity, w io.Writer) error) {
	outWrite(out+".zip", func(writer io.Writer) error {

		zf := zip.NewWriter(writer)
		defer zf.Close()

		var werr error

		report.DepDetailGraph.ForEach(func(n *detail.DepDetailGraph) bool {

			if n.Name == "" {
				return true
			}

			tag, err := swid.NewTag(n.Dep.Key(), n.Name, n.Version)
			if err != nil {
				logs.Warn(err)
				return true
			}

			tag.TagVersion = 1
			tag.SoftwareName = n.Name
			tag.SoftwareVersion = n.Version
			tag.VersionScheme = &swid.VersionScheme{}
			tag.VersionScheme.SetCode(1)

			if n.Vendor != "" {
				e := swid.Entity{
					RegID:      n.Vendor,
					EntityName: "The vendor of component",
					Roles:      swid.Roles{},
				}
				e.Roles.Set("softwareCreator")
				tag.AddEntity(e)
			}

			name := []string{}
			if n.Vendor != "" {
				name = append(name, n.Vendor)
			}
			name = append(name, n.Name)
			if n.Version != "" {
				name = append(name, n.Version)
			}

			w, err := zf.Create(strings.Join(name, "-") + filepath.Ext(out))
			if err != nil {
				logs.Warn(err)
				return true
			}

			werr = errors.Join(werr, writeFunc(tag, w))

			return true
		})

		return werr
	})
}

func SwidJson(report Report, out string) {
	swidZip(out, report, func(tag *swid.SoftwareIdentity, w io.Writer) error {
		return json.NewEncoder(w).Encode(tag)
	})
}

func SwidXml(report Report, out string) {
	swidZip(out, report, func(tag *swid.SoftwareIdentity, w io.Writer) error {
		return xml.NewEncoder(w).Encode(tag)
	})
}
