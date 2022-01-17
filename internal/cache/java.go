/*
 * @Description: 缓存java文件
 * @Date: 2022-01-08 15:49:43
 */

package cache

import (
	"fmt"
	"opensca/internal/srt"
	"path"
)

/**
 * @description: 获取pom缓存路径
 * @param {srt.Dependency} dep 依赖信息
 * @return {string} 缓存路径
 */
func pomPath(dep srt.Dependency) string {
	return path.Join("maven", dep.Vendor, dep.Name, dep.Version.Org, fmt.Sprintf("%s-%s.pom", dep.Name, dep.Version.Org))
}

/**
 * @description: 保存pom文件
 * @param {srt.Dependency} dep 依赖信息
 * @param {[]byte} data pom文件数据
 */
func SavePom(dep srt.Dependency, data []byte) {
	save(pomPath(dep), data)
}

/**
 * @description: 读取pom文件
 * @param {srt.Dependency} dep 依赖信息
 * @return {[]byte} pom文件数据
 */
func LoadPom(dep srt.Dependency) []byte {
	return load(pomPath(dep))
}
