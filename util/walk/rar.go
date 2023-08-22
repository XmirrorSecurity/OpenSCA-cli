package walk

import (
	"io"
	"os"
	"path/filepath"

	"github.com/xmirrorsecurity/opensca-cli/util/logs"

	"github.com/nwaples/rardecode"
)

func xrar(input, output string) bool {

	if !checkFileHead(input, M_RAR) {
		return false
	}

	fr, err := rardecode.OpenReader(input, "")
	if err != nil {
		logs.Warn(err)
		return false
	}
	defer fr.Close()

	for {

		fh, err := fr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			logs.Warn(err)
			continue
		}

		fp := filepath.Join(output, fh.Name)
		if fh.IsDir {
			os.MkdirAll(fp, 0755)
			continue
		}

		fw, err := os.Create(fp)
		if err != nil {
			logs.Warn(err)
			continue
		}

		io.Copy(fw, fr)
		fw.Close()
	}
	return true
}
