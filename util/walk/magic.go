package walk

import (
	"bytes"
	"io"
	"os"
	"strings"
	"util/logs"
)

type Magic []byte

var (
	M_ZIP = Magic{0x50, 0x4B, 0x03, 0x04}
	M_RAR = Magic{0x52, 0x61, 0x72, 0x21, 0x1A, 0x07}
	M_GZ  = Magic{0x1F, 0x8B}
	M_BZ2 = Magic{0x42, 0x5A, 0x68}
	M_LZ4 = Magic{0x04, 0x22, 0x4D, 0x18}
	M_XZ  = Magic{0xFD, 0x37, 0x7A, 0x58, 0x5A, 0x00}
	M_AR  = Magic{0x21, 0x3C, 0x61, 0x72, 0x63, 0x68, 0x3E, 0x0A}
	M_7Z  = Magic{0x37, 0x7A, 0xBC, 0xAF, 0x27, 0x1C}
)

// checkFileExt 检查文件后缀
func checkFileExt(abspath string, exts ...string) bool {
	for _, ext := range exts {
		if ext == "" {
			continue
		}
		if strings.HasSuffix(abspath, ext) {
			return true
		}
	}
	return false
}

// checkFileHead 检查文件头
func checkFileHead(abspath string, ms ...Magic) bool {
	reader, err := os.Open(abspath)
	if err != nil {
		logs.Warn(err)
	}
	defer reader.Close()
	for _, m := range ms {
		if len(m) == 0 {
			continue
		}
		h := make([]byte, len(m))
		reader.Seek(0, io.SeekStart)
		reader.Read(h)
		reader.Seek(0, io.SeekStart)
		if bytes.Equal(m, h) {
			return true
		}
	}
	return false
}
