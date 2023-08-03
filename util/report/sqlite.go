package report

import (
	"database/sql"
	_ "embed"
	"fmt"
	"path/filepath"
	"strings"
	"util/args"
	"util/logs"
	"util/model"

	_ "github.com/glebarez/go-sqlite"
)

//go:embed sqlite_init.sql
var initSql string

// Sqlite sql格式报告数据
func Sqlite(dep *model.DepTree, taskInfo TaskInfo) {

	initRequired := false

	dbFile := args.Config.Out
	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		logs.Error(err)
	}

	if initRequired {
		logs.Info("initing database: " + dbFile)
		db.Exec(initSql)
	}

	if taskInfo.Error != nil {
		taskInfo.ErrorString = taskInfo.Error.Error()
	}

	absPath, _ := filepath.Abs(filepath.Clean(args.Config.Path))
	moduleName := filepath.Base(absPath)

	if dep.Name != "" {
		moduleName = dep.Name
	}

	result := fmt.Sprintf("\n---- sql report of %s\n", moduleName)
	insertFmt := "insert or ignore into component (name, version, vendor, language, purl) values ('%s','%s','%s','%s','%s');\n"
	insertRef := "insert or ignore into reference (module_name, purl) values ('%s','%s');\n"

	// 遍历所有组件树，提取需求字段，输出SQL语句
	q := []*model.DepTree{dep}
	for len(q) > 0 {
		n := q[0]
		q = append(q[1:], n.Children...)

		if n.Name != "" {
			db.Exec("insert or ignore into component (name, version, vendor, language, purl) values (?,?,?,?,?)", n.Name, n.VersionStr, n.Vendor, n.LanguageStr, n.Purl())
			result = result + fmt.Sprintf(insertFmt, quoteEscape(n.Name), quoteEscape(n.VersionStr), quoteEscape(n.Vendor), quoteEscape(n.LanguageStr), quoteEscape(n.Purl()))
		}

	}

	q = []*model.DepTree{dep}
	for len(q) > 0 {
		n := q[0]
		q = append(q[1:], n.Children...)

		if n.Name != "" {
			result = result + fmt.Sprintf(insertRef, moduleName, quoteEscape(n.Purl()))
			db.Exec("insert or ignore into reference (module_name, purl) values (?, ?)", moduleName, n.Purl())
		}

	}

	result = result + fmt.Sprintf("---- sql report of %s\n", moduleName)
	logs.Debug(result)
}

// quoteEscape 转义单引号
func quoteEscape(src string) (out string) {
	out = strings.ReplaceAll(src, `'`, "\\'")
	return out
}
