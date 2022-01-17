/*
 * @Description: 缓存javascript文件
 * @Date: 2022-01-08 15:54:45
 */

package cache

import (
	"fmt"
	"opensca/internal/srt"
	"path"
)

/**
 * @description: 获取npm缓存路径
 * @param {srt.Dependency} dep 依赖信息
 * @return {string} 缓存路径
 */
func npmPath(dep srt.Dependency) string {
	return path.Join("npm", dep.Name,
		fmt.Sprintf("%s.json", dep.Name))
}

/**
 * @description: 缓存npm数据
 * @param {srt.Dependency} dep 依赖信息
 * @param {[]byte} data 文件数据
 */
func SaveNpm(dep srt.Dependency, data []byte) {
	save(npmPath(dep), data)
}

/**
 * @description: 读取npm缓存数据
 * @param {srt.Dependency} dep 依赖信息
 * @return {[]byte} 缓存文件数据
 */
func LoadNpm(dep srt.Dependency) []byte {
	return load(npmPath(dep))
}
