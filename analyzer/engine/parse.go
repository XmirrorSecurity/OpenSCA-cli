/*
 * @Descripation: 解析依赖
 * @Date: 2021-11-16 20:09:17
 */

package engine

import (
	"analyzer/java"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strings"
	"util/client"
	"util/enum/language"
	"util/filter"
	"util/logs"
	"util/model"
)

// copyright匹配优先级
const (
	low = iota
	mid
	high
)

// PackageVersionFile 需要解析包版本的文件数据
type PackageVersionFile struct {
	Language language.Type
	FileData *model.FileInfo
}

// PackageBase package.json、composer.json提取需要的字段
type PackageBase struct {
	PackageName    string `json:"name"`
	PackageVersion string `json:"version"`
}

// parseDependency 解析依赖
func (e Engine) parseDependency(dirRoot *model.DirTree, depRoot *model.DepTree) *model.DepTree {
	if depRoot == nil {
		depRoot = model.NewDepTree(nil)
	}
	var copyrightMess = make(map[string]string)
	var h = md5.New()
	var toHash = false
	var packageVersionFiles []PackageVersionFile
	var packageVersionFileCount = 0
	for _, analyzer := range e.Analyzers {
		// 遍历目录树获取要检测的文件
		files := []*model.FileInfo{}
		q := []*model.DirTree{dirRoot}
		for len(q) > 0 {
			n := q[0]
			q = q[1:]
			for _, dir := range n.DirList {
				q = append(q, n.SubDir[dir])
			}
			for _, f := range n.Files {
				if analyzer.CheckFile(f.Name) {
					files = append(files, f)
					//计算匹配到的文件hash。文件顺序会影响hash值，比如文件重命名后顺序改变
					h.Write(f.Data)
					if !toHash {
						toHash = true
					}
					//只保存根目录下可以版本提取的文件
					if strings.HasSuffix(f.Name, client.PackageBasePath+"/"+path.Base(f.Name)) {
						packageVersionFileCount = packageVersionFileCount + 1
						toParseVersion := false
						//如果不满足条件则可能不是单一语言的包，不提取版本
						switch analyzer.GetLanguage() {
						//java最多只有1个
						case language.Java:
							if packageVersionFileCount == 1 {
								toParseVersion = true
							}
						//php、ruby、go最多有两个
						case language.Php, language.Ruby, language.Golang, language.Rust:
							if packageVersionFileCount <= 2 {
								toParseVersion = true
							}
						//js最多有3个
						case language.JavaScript:
							if packageVersionFileCount <= 3 {
								toParseVersion = true
							}
						//python最多有5个
						case language.Python:
							if packageVersionFileCount <= 5 {
								toParseVersion = true
							}
						}
						if toParseVersion {
							if analyzer.GetLanguage() == language.JavaScript {
								//js只处理package.json
								if !filter.JavaScriptPackage(f.Name) {
									continue
								}
							} else if analyzer.GetLanguage() == language.Php {
								//php只处理composer.json
								if !filter.PhpComposer(f.Name) {
									continue
								}
							} else if analyzer.GetLanguage() == language.Python {
								//python只处理setup.py、pyproject.toml
								if !filter.PythonSetup(f.Name) && !filter.PythonPyproject(f.Name) {
									continue
								}
							} else if analyzer.GetLanguage() == language.Ruby {
								//ruby只处理Gemfile
								if !filter.RubyGemfile(f.Name) {
									continue
								}
							} else if analyzer.GetLanguage() == language.Golang {
								//go只处理go.mod
								if !filter.GoMod(f.Name) {
									continue
								}
							} else if analyzer.GetLanguage() == language.Rust {
								//rust只处理Cargo.toml
								if !filter.RustCargoToml(f.Name) {
									continue
								}
							}
							packageVersionFile := PackageVersionFile{Language: analyzer.GetLanguage(), FileData: f}
							packageVersionFiles = append(packageVersionFiles, packageVersionFile)
						} else {
							packageVersionFiles = nil
						}
					}
				} else if filter.CheckLicense(f.Name) {
					if _, ok := copyrightMess[path.Dir(f.Name)]; !ok {
						// 记录解析到的copyrigh信息
						copyrightMess[path.Dir(f.Name)] = parseCopyright(f)
					}
				}
			}
		}

		// 从文件中解析依赖树
		for _, d := range analyzer.ParseFiles(files) {
			p := path.Dir(d.Path)
			if _, ok := copyrightMess[p]; ok {
				// 将copyright信息加入与其同一文件目录的依赖节点中
				d.CopyrightText = copyrightMess[p]
				delete(copyrightMess, p)
			}
			depRoot.Children = append(depRoot.Children, d)
			d.Parent = depRoot
			if d.Name != "" && !strings.ContainsAny(d.Vendor+d.Name, "${}") && d.Version.Ok() {
				d.Path = path.Join(d.Path, d.Dependency.String())
			}
			// 标识为直接依赖
			d.Direct = true
			for _, c := range d.Children {
				c.Direct = true
			}
			// 填充路径
			q := []*model.DepTree{d}
			s := map[int64]struct{}{}
			for len(q) > 0 {
				n := q[0]
				n.Language = analyzer.GetLanguage()
				if _, ok := s[n.ID]; !ok {
					s[n.ID] = struct{}{}
					for _, c := range n.Children {
						if c.Path == "" {
							// 路径为空的组件在父组件路径后拼接本身依赖信息
							c.Path = path.Join(n.Path, c.Dependency.String())
						} else {
							// 路径不为空的组件在组件路径后拼接本身依赖信息
							c.Path = path.Join(c.Path, c.Dependency.String())
						}
					}
					q = append(q[1:], n.Children...)
				} else {
					q = q[1:]
				}
			}
		}
	}
	if toHash {
		client.PackageHash = base64.StdEncoding.EncodeToString(h.Sum(nil))
	}
	if client.PackageVersion == "" && len(packageVersionFiles) >= 1 {
		parsePackageVersion(packageVersionFiles)
	}
	// 删除依赖树空节点
	q := []*model.DepTree{depRoot}
	for len(q) > 0 {
		n := q[0]
		q = append(q[1:], n.Children...)
		if n.Name == "" || strings.ContainsAny(n.Vendor+n.Name, "${}") || !n.Version.Ok() {
			n.Move(n.Parent)
		}
	}
	// 校对根节点
	if depRoot.Name == "" {
		var d *model.DepTree
		for _, c := range depRoot.Children {
			if !filter.AllPkg(c.Path) {
				if d != nil {
					d = nil
					break
				} else {
					d = c
				}
			}
		}
		if d != nil {
			depRoot.Dependency = d.Dependency
			depRoot.Path = d.Path
			d.Move(depRoot)
		}
	}
	return depRoot
}

// 从文件中提取copyright信息
func parseCopyright(f *model.FileInfo) string {
	matchLevel := map[int]string{}
	ct := string(f.Data)
	if len(ct) == 0 {
		return ""
	}
	pras := strings.Split(ct, "\n\n")
	re := regexp.MustCompile(`^\d{4}$|^\d{4}-\d{4}$|^\(c\)$`)
	for _, pra := range pras {
		if !strings.Contains(strings.ToLower(pra), "copyright") {
			continue
		}
		lines := strings.Split(pra, "\n")
		line := strings.TrimSpace(lines[0])
		if len(lines) == 0 {
			continue
		}
		tks := strings.Fields(line)
		if len(tks) == 0 {
			continue
		}
		if strings.EqualFold("copyright", tks[0]) && len(tks) > 1 {
			if re.MatchString(tks[1]) {
				matchLevel[high] = line
			}
			matchLevel[mid] = line
		}
		for _, l := range lines {
			if strings.HasPrefix(strings.TrimSpace(strings.ToLower(l)), "copyright") {
				matchLevel[low] = strings.TrimSpace(l)
				break
			}
		}

	}
	for i := high; i >= low; i-- {
		if matchLevel[i] != "" {
			return matchLevel[i]
		}
	}
	return ""
}

// 从特定文件中提取包名、版本信息
func parsePackageVersion(packageVersionFiles []PackageVersionFile) {
	for _, packageVersionFile := range packageVersionFiles {
		switch packageVersionFile.Language {
		case language.Java:
			if filter.JavaPom(packageVersionFile.FileData.Name) {
				p := java.ReadPom(packageVersionFile.FileData.Data)
				client.PackageVersion = p.Version
				if p.ArtifactId != "" {
					client.PackageName = p.ArtifactId
				}
			} else if strings.HasSuffix(packageVersionFile.FileData.Name, ".gradle") {
				//基本没有声明包名和版本号？ 暂不解析
			}
		case language.JavaScript, language.Php:
			packageBase := PackageBase{}
			err := json.Unmarshal(packageVersionFile.FileData.Data, &packageBase)
			if err != nil {
				logs.Warn(err)
			} else {
				if packageBase.PackageName != "" {
					client.PackageName = packageBase.PackageName
				}
				if packageBase.PackageVersion != "" {
					client.PackageVersion = packageBase.PackageVersion
				}
			}
		case language.Ruby:
			//根路径是否存在gemspec
			gemspecPath := ""
			rootDir := path.Dir(packageVersionFile.FileData.Name)
			files, err := ioutil.ReadDir(rootDir)
			if err != nil {
				continue
			}
			for _, file := range files {
				if !file.IsDir() {
					if strings.HasSuffix(file.Name(), "gemspec") {
						gemspecPath = path.Join(rootDir, file.Name())
						if data, err := ioutil.ReadFile(gemspecPath); err == nil {
							pkgMatch := regexp.MustCompile(`\s*spec\.name\s*=\s*["'](.*)["']\r?\n`)
							r := pkgMatch.FindSubmatch(data)
							if len(r) == 2 {
								if string(r[1]) != "" {
									client.PackageName = string(r[1])
								}
							}
							pkgVerMatch := regexp.MustCompile(`\s*spec\.version\s*=\s*["'](.*)["']\r?\n`)
							r = pkgVerMatch.FindSubmatch(data)
							if len(r) == 2 {
								if string(r[1]) != "" {
									client.PackageVersion = string(r[1])
								}
							} else {
								pkgVerPathMatch := regexp.MustCompile(`require\s*["'](.*version)["']\r?\n`)
								r = pkgVerPathMatch.FindSubmatch(data)
								if len(r) == 2 {
									if string(r[1]) != "" {
										//一般在lib目录
										pkgVersionPath := fmt.Sprintf("%s/lib/%s.rb", rootDir, string(r[1]))
										if f, err := os.Stat(pkgVersionPath); err == nil {
											if !f.IsDir() {
												if data, err := ioutil.ReadFile(pkgVersionPath); err == nil {
													pkgVerMatch = regexp.MustCompile(`\s*VERSION\s*=\s*["'](.*)["']\r?\n`)
													r = pkgVerMatch.FindSubmatch(data)
													if len(r) == 2 {
														if string(r[1]) != "" {
															client.PackageVersion = string(r[1])
														}
													}
												}
											}
										}
									}
								}
							}
						}
						return
					}
				}
			}
		case language.Golang:
			//只能提取包名
			pkgMatch := regexp.MustCompile(`module\s*(.*)\r?\n`)
			r := pkgMatch.FindSubmatch(packageVersionFile.FileData.Data)
			if len(r) == 2 {
				if string(r[1]) != "" {
					client.PackageName = string(r[1])
				}
			}
		case language.Rust:
			pkgNameMatch := regexp.MustCompile(`\nname\s*=\s*['"](.*)['"]\r?\n`)
			pkgVerMatch := regexp.MustCompile(`\nversion\s*=\s*['"](.*)['"]\r?\n`)
			r := pkgNameMatch.FindSubmatch(packageVersionFile.FileData.Data)
			if len(r) == 2 {
				if string(r[1]) != "" {
					client.PackageName = string(r[1])
				}
			}
			r = pkgVerMatch.FindSubmatch(packageVersionFile.FileData.Data)
			if len(r) == 2 {
				if string(r[1]) != "" {
					client.PackageVersion = string(r[1])
					return
				}
			}
		case language.Python:
			//先从特定文件中尝试提取包名和版本
			for _, filepath := range []string{
				path.Dir(packageVersionFile.FileData.Name) + "/PKG-INFO",
				path.Dir(packageVersionFile.FileData.Name) + "/METADATA",
			} {
				if f, err := os.Stat(filepath); err == nil {
					if !f.IsDir() {
						if data, err := ioutil.ReadFile(filepath); err == nil {
							pkgNameMatch := regexp.MustCompile(`\nName: (.*)\r?\n`)
							r := pkgNameMatch.FindSubmatch(data)
							if len(r) == 2 {
								if string(r[1]) != "" {
									client.PackageName = string(r[1])
								}
							}
							pkgVerMatch := regexp.MustCompile(`\nVersion: (.*)\r?\n`)
							r = pkgVerMatch.FindSubmatch(data)
							if len(r) == 2 {
								if string(r[1]) != "" {
									client.PackageVersion = string(r[1])
									return
								}
							}
						}
					}
				}
			}
			pkgVerMatch := &regexp.Regexp{}
			//可能存在name和version不是相邻的，暂时不考虑
			if filter.PythonSetup(packageVersionFile.FileData.Name) {
				pkgVerMatch = regexp.MustCompile(`setup\(\s*name\s*=\s*["'](.*)["'],\s*version\s*=\s*["'](.*)["'],`)
			} else if filter.PythonPyproject(packageVersionFile.FileData.Name) {
				pkgVerMatch = regexp.MustCompile(`\[tool.poetry\]\s*name\s*=\s*"(.*)"\s*version\s*=\s*"(.*)"\r?\n`)
			}
			if pkgVerMatch.String() != "" {
				r := pkgVerMatch.FindSubmatch(packageVersionFile.FileData.Data)
				if len(r) == 3 {
					if string(r[1]) != "" {
						client.PackageName = string(r[1])
					}
					if string(r[2]) != "" {
						client.PackageVersion = string(r[2])
					}
				}
			}
		}
	}
}
