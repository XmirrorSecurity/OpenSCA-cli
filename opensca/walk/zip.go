package walk

import (
	"archive/zip"
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/xmirrorsecurity/opensca-cli/v3/opensca/common"
	"github.com/xmirrorsecurity/opensca-cli/v3/opensca/logs"

	"github.com/axgle/mahonia"
)

func xzip(ctx context.Context, filter ExtractFileFilter, input, output string) bool {

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

		select {
		case <-ctx.Done():
			return false
		default:
		}

		if f.FileInfo().IsDir() {
			continue
		}

		fp := filepath.Join(output, f.Name)

		if f.Flags == 0 {
			gbk := mahonia.NewDecoder("gbk").ConvertString(f.Name)
			_, cdata, _ := mahonia.NewDecoder("utf-8").Translate([]byte(gbk), true)
			fp = filepath.Join(output, string(cdata))
		}

		// avoid zip slip
		if !strings.HasPrefix(fp, filepath.Clean(output)+string(os.PathSeparator)) {
			logs.Warn("Invalid file path: %s", fp)
			continue
		}

		if filter != nil && !filter(fp) {
			continue
		}

		fr, err := f.Open()
		if err != nil {
			logs.Warn(err)
			continue
		}

		os.MkdirAll(filepath.Dir(fp), 0777)
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

func xjar(ctx context.Context, filter ExtractFileFilter, input, output string) bool {

	if !checkFileExt(input, ".jar") {
		return false
	}

	// 生成临时文件
	tempf := common.CreateTemp("jar")
	defer os.Remove(tempf.Name())

	data, err := io.ReadAll(tempf)
	if err != nil {
		logs.Warn(err)
		tempf.Close()
		return false
	}

	// 剔除可执行jar包前的bash脚本
	i := bytes.Index(data, M_ZIP)
	if i == -1 {
		tempf.Close()
		return false
	}

	tempf.Write(data[i:])
	tempf.Close()

	return xzip(ctx, filter, tempf.Name(), output)
}
