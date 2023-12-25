package walk

import (
	"archive/tar"
	"compress/bzip2"
	"compress/gzip"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/xmirrorsecurity/opensca-cli/v3/opensca/logs"
)

func xtar(ctx context.Context, filter ExtractFileFilter, input, output string) bool {

	if !checkFileExt(input, ".tar") {
		return false
	}

	f, err := os.Open(input)
	if err != nil {
		logs.Warn(err)
		return false
	}
	defer f.Close()

	fr := tar.NewReader(f)
	for {

		select {
		case <-ctx.Done():
			return false
		default:
		}

		fh, err := fr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			logs.Warn(err)
			break
		}

		fp := filepath.Join(output, fh.Name)

		// avoid zip slip
		if !strings.HasPrefix(fp, filepath.Clean(output)+string(os.PathSeparator)) {
			logs.Warn("Invalid file path: %s", fp)
			continue
		}

		if fh.Typeflag == tar.TypeDir {
			os.MkdirAll(fp, 0755)
			continue
		}

		if filter != nil && !filter(fp) {
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

func xgz(input, output string) bool {

	if !checkFileHead(input, M_GZ) {
		return false
	}

	f, err := os.Open(input)
	if err != nil {
		logs.Warn(err)
		return false
	}
	defer f.Close()

	fr, err := gzip.NewReader(f)
	if err != nil {
		logs.Warn(err)
		return false
	}
	defer fr.Close()

	fp := filepath.Join(output, strings.TrimSuffix(filepath.Base(input), filepath.Ext(input)))
	os.MkdirAll(filepath.Dir(fp), 0777)
	fw, err := os.Create(fp)
	if err != nil {
		logs.Warn(err)
		return false
	}

	_, err = io.Copy(fw, fr)
	fw.Close()

	return err == nil
}

func xbz2(input, output string) bool {

	if !checkFileHead(input, M_BZ2) {
		return false
	}

	f, err := os.Open(input)
	if err != nil {
		logs.Warn(err)
		return false
	}
	defer f.Close()

	fr := bzip2.NewReader(f)

	fp := filepath.Join(output, strings.TrimSuffix(filepath.Base(input), filepath.Ext(input)))
	fw, err := os.Create(fp)
	if err != nil {
		logs.Warn(err)
		return false
	}

	_, err = io.Copy(fw, fr)
	fw.Close()

	return err == nil
}
