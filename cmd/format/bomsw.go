package format

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"sort"
	"strings"

	"github.com/xmirrorsecurity/opensca-cli/v3/cmd/detail"
	"github.com/xmirrorsecurity/opensca-cli/v3/opensca/model"
)

func BomSWJson(report Report, out string) {
	outWrite(out, func(w io.Writer) error {
		doc := bomSWDoc(report)
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		return encoder.Encode(doc)
	})
}

func bomSWDoc(report Report) *model.BomSWDocument {

	doc := model.NewBomSWDocument(report.TaskInfo.AppName, "opensca-cli")
	defer func() {
		doc.SbomHashCheck = calculateSbomHashCheck(doc)
	}()

	report.DepDetailGraph.ForEach(func(n *detail.DepDetailGraph) bool {

		if n.Name == "" {
			return true
		}

		lics := []string{}
		for _, lic := range n.Licenses {
			lics = append(lics, lic.ShortName)
		}
		doc.AppendComponents(func(swc *model.BomSWComponent) {
			swc.ID = n.Purl()
			swc.Name = n.Name
			swc.Version = n.Version
			swc.License = lics
		})

		children := []string{}
		for _, c := range n.Children {
			if c.Name == "" {
				continue
			}
			children = append(children, c.Purl())
		}
		doc.AppendDependencies(n.Purl(), children)

		return true
	})

	return doc
}

func calculateSbomHashCheck(doc *model.BomSWDocument) string {
	sha256Hash := sha256.New()
	writeHash := func(v string) { sha256Hash.Write([]byte(v)) }
	// basic info
	writeHash(doc.Basic.DocumentName)
	writeHash(doc.Basic.DocumentVersion)
	writeHash(doc.Basic.DocumentTime)
	writeHash(doc.Basic.SbomFormat)
	writeHash(doc.Basic.ToolInfo)
	writeHash(doc.Basic.SbomAuthor)
	writeHash(doc.Basic.SbomAuthorComments)
	writeHash(doc.Basic.SbomComments)
	// components
	for _, component := range doc.Software.Components {
		writeHash(sortMapString(component.Author))
		writeHash(sortMapString(component.Provider))
		writeHash(component.Name)
		writeHash(component.Version)
		for _, hash := range component.HashValue {
			writeHash(hash.Algorithm + ":" + hash.Value)
		}
		writeHash(component.ID)
		for _, lic := range component.License {
			writeHash(lic)
		}
		writeHash(component.Timestamp)
	}
	// dependencies
	for _, deps := range doc.Software.Dependencies {
		writeHash(deps.Ref)
		for _, depsOn := range deps.DependsOn {
			writeHash(depsOn.Ref)
		}
	}
	hashStr := hex.EncodeToString(sha256Hash.Sum(nil)[:])
	return hashStr
}

func sortMapString(v map[string]string) string {
	keys := []string{}
	for key := range v {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	res := strings.Builder{}
	for _, key := range keys {
		res.WriteString(key + ":" + v[key])
	}
	return res.String()
}
