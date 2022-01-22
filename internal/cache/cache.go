/*
 * @Description: cache download file
 * @Date: 2022-01-08 15:34:37
 */

package cache

import (
	"fmt"
	"io/ioutil"
	"opensca/internal/args"
	"opensca/internal/enum/language"
	"opensca/internal/logs"
	"opensca/internal/srt"
	"os"
	"path"
	"strings"
)

var cacheDir string

func init() {
	// create cache dir
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
 * @description: save cache file
 * @param {string} filepath cache file path
 * @param {[]byte} cache file data
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
 * @description: load cache file
 * @param {string} filepath cache file path
 * @return {[]byte} cache file data
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

func filepath(dep srt.Dependency) string {
	switch dep.Language {
	case language.Java:
		return path.Join("maven", dep.Vendor, dep.Name, dep.Version.Org, fmt.Sprintf("%s-%s.pom", dep.Name, dep.Version.Org))
	case language.JavaScript:
		return path.Join("npm", fmt.Sprintf("%s.json", dep.Name))
	case language.Php:
		return path.Join("composer", fmt.Sprintf("%s.json", dep.Name))
	default:
		return path.Join("none", fmt.Sprintf("%s-%s-%s", dep.Vendor, dep.Name, dep.Version.Org))
	}
}

/**
 * @description: save cache file
 * @param {srt.Dependency} dep dependency infomation
 * @param {[]byte} data cache file data
 */
func SaveCache(dep srt.Dependency, data []byte) {
	save(filepath(dep), data)
}

/**
 * @description: load cache file
 * @param {srt.Dependency} dep dependency infomation
 * @return {[]byte} cache file data
 */
func LoadCache(dep srt.Dependency) []byte {
	return load(filepath(dep))
}
