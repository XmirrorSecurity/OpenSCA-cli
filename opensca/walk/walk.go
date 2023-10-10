package walk

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/xmirrorsecurity/opensca-cli/opensca/common"
	"github.com/xmirrorsecurity/opensca-cli/opensca/logs"
	"github.com/xmirrorsecurity/opensca-cli/opensca/model"
	"github.com/xmirrorsecurity/opensca-cli/opensca/sca/filter"
)

var (
	wg = sync.WaitGroup{}
)

type ExtractFileFilter func(relpath string) bool
type WalkFileFunc func(parent *model.File, files []*model.File)

// Walk 遍历文件/目录/压缩包
// name: 检测文件名
// origin: 检测数据源
// filter: 过滤需要提取的文件
// do: 对文件的操作
// size: 检测文件大小
func Walk(ctx context.Context, name, origin string, filter ExtractFileFilter, do WalkFileFunc) (size int64, err error) {

	defer wg.Wait()

	delete, filepath, err := download(origin)
	if err != nil {
		return
	}

	if delete {
		defer os.RemoveAll(filepath)
	}

	if f, xerr := os.Stat(filepath); xerr == nil {
		if !f.IsDir() {
			size = f.Size()
		}
	} else {
		logs.Error(xerr)
		err = xerr
		return
	}

	parent := model.NewFile(filepath, name)
	err = walk(ctx, parent, filter, do)
	return
}

func walk(ctx context.Context, parent *model.File, filterFunc ExtractFileFilter, walkFunc WalkFileFunc) error {

	var files []*model.File

	err := filepath.Walk(parent.Abspath(), func(path string, info fs.FileInfo, err error) error {

		select {
		case <-ctx.Done():
			return nil
		default:
		}

		if err != nil {
			logs.Warn(err)
			return nil
		}
		if info.IsDir() {
			if strings.HasSuffix(path, ".git") || strings.HasSuffix(path, ".opensca-cache") || strings.HasSuffix(path, ".temp") {
				return filepath.SkipDir
			}
			return nil
		}

		rel := filepath.Join(parent.Relpath(), strings.TrimPrefix(path, parent.Abspath()))

		if filterFunc != nil && !filterFunc(rel) {
			return nil
		}

		if !filter.CompressFile(rel) {
			logs.Debugf("find %s", rel)
			files = append(files, model.NewFile(path, rel))
			return nil
		}

		decompress(ctx, path, filterFunc, func(dir string) {
			wg.Add(1)
			go func() {
				defer wg.Done()
				defer os.RemoveAll(dir)
				parent := model.NewFile(dir, rel)
				if err := walk(ctx, parent, filterFunc, walkFunc); err != nil {
					logs.Warn(err)
				}
			}()
		})

		return nil
	})

	walkFunc(parent, files)
	return err
}

// decompress 解压到指定位置
// input: 压缩包绝对路径
// do: 对解压后目录的操作
// do.tmpdir: 临时解压目录绝对路径 需要手动删除目录
func decompress(ctx context.Context, input string, filter ExtractFileFilter, do func(tmpdir string)) {
	tmp := common.MkdirTemp("decompress")
	ok := false ||
		xzip(ctx, filter, input, tmp) ||
		xjar(ctx, filter, input, tmp) ||
		xrar(ctx, filter, input, tmp) ||
		xtar(ctx, filter, input, tmp) ||
		xgz(input, tmp) ||
		xbz2(input, tmp) ||
		false
	if ok {
		do(tmp)
	} else {
		os.RemoveAll(tmp)
	}
}
