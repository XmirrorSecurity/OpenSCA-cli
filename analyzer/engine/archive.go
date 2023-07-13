/*
 * @Description: 解压相关操作
 * @Date: 2021-11-04 16:13:57
 */

package engine

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"util/bar"
	"util/filter"
	"util/logs"
	"util/model"
	"util/temp"

	"github.com/axgle/mahonia"
	"github.com/mholt/archiver"
	"github.com/nwaples/rardecode"
	"github.com/pkg/errors"
)

// checkFile 检测是否为可检测的文件
func (e Engine) checkFile(filename string) bool {
	for _, analyzer := range e.Analyzers {
		if analyzer.CheckFile(filename) ||
			filter.CheckLicense(filename) {
			return true
		}
	}
	return false
}

// unArchiveFile 解压文件获取目录树
func (e Engine) unArchiveFile(filepath string) (root *model.DirTree) {
	filepath = strings.ReplaceAll(filepath, `\`, `/`)
	// 目录树根
	root = model.NewDirTree()
	var walker archiver.Walker
	if filter.Tar(filepath) {
		walker = archiver.NewTar()
	} else if filter.Zip(filepath) || filter.Jar(filepath) {
		walker = archiver.NewZip()
	} else if filter.Rar(filepath) {
		walker = archiver.NewRar()
	} else if filter.TarGz(filepath) {
		walker = archiver.NewTarGz()
	} else if filter.TarBz2(filepath) {
		walker = archiver.NewTarBz2()
	}
	if err := walker.Walk(filepath, func(f archiver.File) error {
		if !f.IsDir() {
			bar.Archive.Add(1)
			// 跳过过大文件
			if f.Size() > 1024*1024*1024 {
				return nil
			}
			fileName := f.Name()
			if file, ok := f.Header.(zip.FileHeader); ok {
				if file.Flags == 0 {
					gbkName := mahonia.NewDecoder("gbk").ConvertString(file.Name)
					_, cdata, _ := mahonia.NewDecoder("utf-8").Translate([]byte(gbkName), true)
					fileName = string(cdata)
				} else {
					fileName = file.Name
				}
			} else if file, ok := f.Header.(*rardecode.FileHeader); ok {
				fileName = file.Name
			} else if file, ok := f.Header.(*gzip.Header); ok {
				fileName = file.Name
			} else if file, ok := f.Header.(*tar.Header); ok {
				fileName = file.Name
			}
			// 读取文件数据
			data, err := ioutil.ReadAll(f)
			if err != nil {
				return errors.WithStack(err)
			}
			// 格式化路径
			fileName = strings.ReplaceAll(fileName, `\`, `/`)
			if e.checkFile(fileName) {
				// 支持解析的文件
				root.AddFile(model.NewFileData(fileName, data))
			} else if filter.AllPkg(fileName) {
				// 将压缩包解压到本地
				temp.DoInTempDir(func(tempdir string) {
					targetPath := path.Join(tempdir, path.Base(fileName))
					if out, err := os.Create(targetPath); err == nil {
						_, err = out.Write(data)
						out.Close()
						if err != nil {
							logs.Error(err)
						}
						// 获取当前目录树
						dir := root.GetDir(fileName)
						name := path.Base(fileName)
						if _, ok := dir.SubDir[name]; !ok {
							// 将压缩包的内容添加到当前目录树
							dir.DirList = append(dir.DirList, name)
							dir.SubDir[name] = e.unArchiveFile(targetPath)
						}
					} else {
						logs.Error(err)
					}
				})
			}
		}
		return nil
	}); err != nil {
		logs.Error(errors.WithMessage(err, filepath))
	}
	return root
}

// opendir 读取目录获取目录树
func (e Engine) opendir(dirpath string) (dir *model.DirTree) {
	bar.Dir.Add(1)
	dir = model.NewDirTree()
	files, err := ioutil.ReadDir(dirpath)
	if err != nil {
		logs.Error(errors.WithMessage(err, dirpath))
		return
	}
	for _, file := range files {
		filename := file.Name()
		filepath := path.Join(dirpath, filename)
		if file.IsDir() {
			dir.DirList = append(dir.DirList, filename)
			dir.SubDir[filename] = e.opendir(filepath)
		} else {
			if filter.AllPkg(filename) {
				dir.DirList = append(dir.DirList, filename)
				dir.SubDir[filename] = e.unArchiveFile(filepath)
			} else if e.checkFile(filename) {
				if data, err := ioutil.ReadFile(filepath); err != nil {
					logs.Error(err)
				} else {
					dir.Files = append(dir.Files, model.NewFileData(filepath, data))
				}
			}
		}
	}
	return
}
