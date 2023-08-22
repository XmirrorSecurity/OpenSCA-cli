package walk

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/xmirrorsecurity/opensca-cli/util/logs"
)

// Walk 遍历文件/目录/压缩包
// relpath: 检测位置的相对路径
// abspath: 检测位置的绝对路径
// do: 目标内文件操作
// do.parent: 文件的上层压缩包
// do.abspath: 文件的绝对路径
// do.relpath: 文件的相对路径
func Walk(relpath, abspath string, do func(parent, abspath, relpath string)) error {
	wg := sync.WaitGroup{}
	defer wg.Wait()

	return filepath.Walk(abspath, func(path string, info fs.FileInfo, err error) error {

		if err != nil {
			logs.Warn(err)
			return nil
		}
		if info.IsDir() {
			return nil
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			Decompress(path, func(tmpdir string) {
				if err := Walk(path, tmpdir, do); err != nil {
					logs.Warn(err)
				}
			})
		}()

		rel := filepath.Join(relpath, strings.TrimPrefix(path, abspath))
		do(relpath, path, rel)

		return nil
	})
}

var (
	tempdir = ".temp"
)

func init() {
	excpath, _ := os.Executable()
	tempdir = filepath.Join(filepath.Dir(excpath), tempdir)
	os.RemoveAll(tempdir)
	os.MkdirAll(tempdir, 0755)
}

// Decompress 解压到指定位置
// input: 压缩包绝对路径
// do: 对解压后目录的操作
// do.tmpdir: 临时解压目录绝对路径
func Decompress(input string, do func(tmpdir string)) {
	// TODO: 临时目录位置改为本地 启动cli时清空
	tmp, _ := os.MkdirTemp(tempdir, "decompress")
	defer os.RemoveAll(tmp)
	ok := false ||
		xzip(input, tmp) ||
		xjar(input, tmp) ||
		xrar(input, tmp) ||
		xtar(input, tmp) ||
		xgz(input, tmp) ||
		xbz2(input, tmp) ||
		false
	if ok {
		do(tmp)
	}
}
