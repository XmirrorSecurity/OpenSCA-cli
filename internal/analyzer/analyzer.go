/*
 * @Descripation: Analyzer接口
 * @Date: 2021-11-17 21:26:36
 */

package analyzer

import (
	"opensca/internal/enum/language"
	"opensca/internal/srt"
)

type Analyzer interface {

	/**
	 * @description: 获取当前Analyzer的语言
	 * @return {language.Type} 语言
	 */
	GetLanguage() language.Type

	/**
	 * @description: 检测是否是可解析的文件
	 * @param {string} filename 文件名
	 * @return {bool} 是可解析的文件返回true
	 */
	CheckFile(filename string) bool

	/**
	 * @descriptsrt筛选当前解析器需要解析的文件
	 * @param {*modsrtrTree} dirRoot 目录树节点
	 * @param {*srt.DepTree} depRoot 依赖树节点
	 * @return {[]*srt.FileData} 需要解析的文件列表
	 */
	FilterFile(dirRoot *srt.DirTree, depRoot *srt.DepTree) []*srt.FileData

	/**
	 * @descriptsrt解析文件
	 * @param {*srt.DirTree} dirRoot 目录树节点
	 * @param {*modsrtpTree} depRoot 依赖树节点
	 * @param {*srt.FileData} file 文件信息
	 * @return {[]*srt.DepTree} 解析出的依赖列表
	 */
	ParseFile(dirRoot *srt.DirTree, depRoot *srt.DepTree, file *srt.FileData) []*srt.DepTree
}
