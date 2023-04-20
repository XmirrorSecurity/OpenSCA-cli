package report

import (
	"archive/zip"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
	"util/logs"
	"util/model"

	"github.com/veraison/swid"
)

func buildSwid(ext string, writer io.Writer, dep *model.DepTree, taskInfo TaskInfo) {
	w := zip.NewWriter(writer)
	defer w.Close()
	q := []*model.DepTree{dep}
	for len(q) > 0 {
		n := q[0]
		q = append(q[1:], n.Children...)
		if n.Name == "" {
			continue
		}
		tag, err := swid.NewTag(fmt.Sprint(n.ID), n.Name, n.GetVersion())
		if err != nil {
			logs.Warn(err)
			continue
		}
		tag.TagVersion = 1
		tag.SoftwareName = n.Name
		tag.SoftwareVersion = n.GetVersion()
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

		name := []string{fmt.Sprint(n.ID)}
		if n.Vendor != "" {
			name = append(name, n.Vendor)
		}
		name = append(name, n.Name)
		if n.GetVersion() != "" {
			name = append(name, n.GetVersion())
		}
		f, err := w.Create(strings.Join(name, "-") + "." + strings.TrimLeft(ext, "."))
		if err != nil {
			logs.Warn(err)
			continue
		}
		if strings.Contains(ext, "json") {
			if err = json.NewEncoder(f).Encode(tag); err != nil {
				logs.Warn(err)
			}
		} else if strings.Contains(ext, "xml") {
			if err = xml.NewEncoder(f).Encode(tag); err != nil {
				logs.Warn(err)
			}
		}
	}
}

func SwidJson(writer io.Writer, dep *model.DepTree, taskInfo TaskInfo) {
	buildSwid("json", writer, dep, taskInfo)
}

func SwidXml(writer io.Writer, dep *model.DepTree, taskInfo TaskInfo) {
	buildSwid("xml", writer, dep, taskInfo)
}
