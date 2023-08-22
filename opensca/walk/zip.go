package walk

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"

	"github.com/xmirrorsecurity/opensca-cli/util/logs"

	"github.com/axgle/mahonia"
)

func xzip(filter ExtractFileFilter, input, output string) bool {
	if !checkFileHead(input, M_ZIP) {
		return false
	}
	rf, err := zip.OpenReader(input)
	if err != nil {
		logs.Warn(err)
		return false
	}
	defer rf.Close()
	for _, f := range rf.File {

		fp := filepath.Join(output, f.Name)

		if f.Flags == 0 {
			gbk := mahonia.NewDecoder("gbk").ConvertString(f.Name)
			_, cdata, _ := mahonia.NewDecoder("utf-8").Translate([]byte(gbk), true)
			fp = filepath.Join(output, string(cdata))
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(fp, 0755); err != nil {
				logs.Warn(err)
			}
			continue
		}

		fr, err := f.Open()
		if err != nil {
			logs.Warn(err)
			continue
		}

		fw, err := os.Create(fp)
		if err != nil {
			logs.Warn(err)
			fr.Close() // 提前退出时手动关闭
			continue
		}

		_, err = io.Copy(fw, fr)
		if err != nil {
			logs.Warn(err)
		}

		fw.Close()
		fr.Close()
	}
	return true
}

func xjar(filter ExtractFileFilter, input, output string) bool {
	if !checkFileExt(input, ".jar") {
		return false
	}
	// TODO: 剔除可执行jar包前的bash脚本
	return xzip(filter, input, output)
}
