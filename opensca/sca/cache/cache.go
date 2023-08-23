package cache

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/xmirrorsecurity/opensca-cli/opensca/logs"
	"github.com/xmirrorsecurity/opensca-cli/opensca/model"
)

var cacheDir string = ".opensca-cache"

func init() {
	excpath, _ := os.Executable()
	cacheDir = filepath.Join(filepath.Dir(excpath), cacheDir)
	if err := os.MkdirAll(cacheDir, 0777); err != nil {
		logs.Error(err)
	}
}

func Save(path string, reader io.Reader) bool {
	os.MkdirAll(filepath.Dir(path), 0777)
	if f, err := os.Create(path); err == nil {
		defer f.Close()
		io.Copy(f, reader)
		return true
	}
	return false
}

func Load(path string, do func(reader io.Reader)) bool {
	if f, err := os.Open(path); err == nil {
		defer f.Close()
		do(f)
		return true
	}
	return false
}

func Path(vendor, name, version string, language model.Language) string {
	var path string
	switch language {
	case model.Lan_Java:
		path = filepath.Join("maven", vendor, name, version, fmt.Sprintf("%s-%s.pom", name, version))
	case model.Lan_JavaScript:
		path = filepath.Join("npm", fmt.Sprintf("%s-%s.json", name, version))
	case model.Lan_Php:
		path = filepath.Join("composer", fmt.Sprintf("%s-%s.json", name, version))
	default:
		path = filepath.Join("none", fmt.Sprintf("%s-%s-%s", vendor, name, version))
	}
	return path
}
