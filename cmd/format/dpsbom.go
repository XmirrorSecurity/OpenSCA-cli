package format

import (
	"archive/zip"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"hash"
	"io"
	"path/filepath"
	"strings"

	"github.com/xmirrorsecurity/opensca-cli/v3/cmd/detail"
	"github.com/xmirrorsecurity/opensca-cli/v3/opensca/model"
)

func DpSbomZip(report Report, out string) {
	zipFile := out
	if !strings.HasSuffix(out, ".zip") {
		zipFile = out + ".zip"
	}
	jsonName := filepath.Base(out)
	if !strings.HasSuffix(jsonName, ".json") {
		jsonName = jsonName + ".json"
	}
	outWrite(zipFile, func(w io.Writer) error {
		doc := pdSbomDoc(report)
		if doc.Hashes.HashFile == "" {
			return errors.New("hash file is required")
		}

		var h hash.Hash
		switch strings.ToLower(doc.Hashes.Algorithm) {
		case "sha-256":
			h = sha256.New()
		case "sha-1":
			h = sha1.New()
		case "md5":
			h = md5.New()
		case "":
			return errors.New("hash algorithm is required")
		default:
			return fmt.Errorf("unsupported hash algorithm: %s", doc.Hashes.Algorithm)
		}

		tojson := func(w io.Writer) error {
			encoder := json.NewEncoder(w)
			encoder.SetIndent("", "  ")
			return encoder.Encode(doc)
		}

		zipfile := zip.NewWriter(w)
		defer zipfile.Close()

		sbomfile, err := zipfile.Create(jsonName)
		if err != nil {
			return err
		}
		err = tojson(sbomfile)
		if err != nil {
			return err
		}

		hashfile, err := zipfile.Create(doc.Hashes.HashFile)
		if err != nil {
			return err
		}
		err = tojson(h)
		if err != nil {
			return err
		}
		hashstr := hex.EncodeToString(h.Sum(nil)[:])
		hashfile.Write([]byte(hashstr))

		return nil
	})
}

func pdSbomDoc(report Report) *model.DpSbomDocument {

	doc := model.NewDpSbomDocument(report.TaskInfo.AppName, "opensca-cli")

	report.DepDetailGraph.ForEach(func(n *detail.DepDetailGraph) bool {

		if n.Name == "" {
			return true
		}

		lics := []string{}
		for _, lic := range n.Licenses {
			lics = append(lics, lic.ShortName)
		}
		doc.AppendComponents(func(dsp *model.DpSbomPackage) {
			dsp.Identifier.Purl = n.Purl()
			dsp.Name = n.Name
			dsp.Version = n.Version
			dsp.License = lics
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
