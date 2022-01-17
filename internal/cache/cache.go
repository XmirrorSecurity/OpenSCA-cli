/*
 * @Description: 缓存下载的文件
 * @Date: 2022-01-08 15:34:37
 */

package cache

import (
	"io/ioutil"
	"opensca/internal/args"
	"opensca/internal/logs"
	"os"
	"path"
	"strings"
)

var cacheDir string

func init() {
	// 创建缓存目录
	cacheDir = ".cache"
	if pwd, err := os.Executable(); err == nil {
		pwd = path.Dir(strings.ReplaceAll(pwd, `\`, `/`))
		cacheDir = path.Join(pwd, ".cache")
	}
	if err := os.MkdirAll(cacheDir, os.ModeDir); err != nil {
		logs.Error(err)
	}
}

/**
 * @description: 保存文件
 * @param {string} filepath 文件保存路径
 * @param {[]byte} data 文件数据
 */
func save(filepath string, data []byte) {
	if args.Cache {
		if err := os.MkdirAll(path.Join(cacheDir, path.Dir(filepath)), os.ModeDir); err == nil {
			if f, err := os.Create(path.Join(cacheDir, filepath)); err == nil {
				defer f.Close()
				f.Write(data)
			}
		}
	}
}

/**
 * @description: 读取文件
 * @param {string} filepath 文件保存路径
 * @return {[]byte} 文件数据
 */
func load(filepath string) []byte {
	if args.Cache {
		if data, err := ioutil.ReadFile(path.Join(cacheDir, filepath)); err == nil {
			return data
		} else {
			return nil
		}
	}
	return []byte{}
}
