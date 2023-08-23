package walk

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/xmirrorsecurity/opensca-cli/opensca/model"
	"github.com/xmirrorsecurity/opensca-cli/util/logs"
)

var (
	tempdir = ".temp"
	wg      = sync.WaitGroup{}
)

func init() {
	excpath, _ := os.Executable()
	tempdir = filepath.Join(filepath.Dir(excpath), tempdir)
	os.RemoveAll(tempdir)
	os.MkdirAll(tempdir, 0755)
}

type ExtractFileFilter func(relpath string) bool
type WalkFileFunc func(parent *model.File, files []*model.File)

// Walk 遍历文件/目录/压缩包
// name: 检测文件名
// origin: 检测数据源
// filter: 过滤需要提取的文件
// do: 对文件的操作
// size: 检测文件大小
func Walk(ctx context.Context, name, origin string, filter ExtractFileFilter, do WalkFileFunc) (err error) {

	defer wg.Wait()

	delete, filepath, err := download(origin)
	if err != nil {
		return
	}

	if delete {
		defer os.RemoveAll(filepath)
	}

	parent := &model.File{Relpath: name, Abspath: filepath}
	err = walk(ctx, parent, filter, do)
	return
}

func walk(ctx context.Context, parent *model.File, filter ExtractFileFilter, do WalkFileFunc) error {

	var files []*model.File

	err := filepath.Walk(parent.Abspath, func(path string, info fs.FileInfo, err error) error {

		if err != nil {
			logs.Warn(err)
			return nil
		}
		if info.IsDir() {
			return nil
		}

		rel := filepath.Join(parent.Relpath, strings.TrimPrefix(path, parent.Abspath))
		files = append(files, &model.File{Abspath: path, Relpath: rel})

		wg.Add(1)
		go func() {
			defer wg.Done()
			decompress(path, filter, func(tmpdir string) {
				parent := &model.File{Relpath: rel, Abspath: tempdir}
				if err := walk(ctx, parent, filter, do); err != nil {
					logs.Warn(err)
				}
			})
		}()

		return nil
	})

	do(parent, files)
	return err
}

// decompress 解压到指定位置
// input: 压缩包绝对路径
// do: 对解压后目录的操作
// do.tmpdir: 临时解压目录绝对路径
func decompress(input string, filter ExtractFileFilter, do func(tmpdir string)) {
	tmp, _ := os.MkdirTemp(tempdir, "decompress")
	defer os.RemoveAll(tmp)
	ok := false ||
		xzip(filter, input, tmp) ||
		xjar(filter, input, tmp) ||
		xrar(filter, input, tmp) ||
		xtar(filter, input, tmp) ||
		xgz(input, tmp) ||
		xbz2(input, tmp) ||
		false
	if ok {
		do(tmp)
	}
}
