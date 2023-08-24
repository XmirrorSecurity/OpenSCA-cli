package report

import (
	"database/sql"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xmirrorsecurity/opensca-cli/opensca/logs"
	"github.com/xmirrorsecurity/opensca-cli/util/args"
	"github.com/xmirrorsecurity/opensca-cli/util/model"

	_ "github.com/glebarez/go-sqlite"
)

//go:embed sqlite_init.sql
var initSql string

// Sqlite sql格式报告数据
func Sqlite(dep *model.DepTree, taskInfo TaskInfo) {

	dbFile := args.Config.Out
	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		logs.Error(err)
	}
	defer db.Close()

	if _, err := os.Stat(dbFile); err != nil {
		logs.Info("initing database: " + dbFile)
		_, err = db.Exec(initSql)
		if err != nil {
			logs.Warn(err)
		}
	}

	if taskInfo.Error != nil {
		taskInfo.ErrorString = taskInfo.Error.Error()
	}

	absPath, _ := filepath.Abs(filepath.Clean(args.Config.Path))
	moduleName := filepath.Base(absPath)

	if dep.Name != "" {
		moduleName = dep.Name
	}

	logs.Debug(fmt.Sprintf("---- sql report of %s", moduleName))
	insertFmt := "insert or ignore into component (name, version, vendor, language, purl) values ('%s','%s','%s','%s','%s');\n"
	insertRef := "insert or ignore into reference (module_name, purl) values ('%s','%s');\n"

	// 遍历所有组件树，提取需求字段，输出SQL语句
	q := []*model.DepTree{dep}
	for len(q) > 0 {
		n := q[0]
		q = append(q[1:], n.Children...)

		if n.Name != "" {
			db.Exec("insert or ignore into component (name, version, vendor, language, purl) values (?,?,?,?,?)", n.Name, n.VersionStr, n.Vendor, n.LanguageStr, n.Purl())
			logs.Debug(fmt.Sprintf(insertFmt, quoteEscape(n.Name), quoteEscape(n.VersionStr), quoteEscape(n.Vendor), quoteEscape(n.LanguageStr), quoteEscape(n.Purl())))
		}

	}

	q = []*model.DepTree{dep}
	for len(q) > 0 {
		n := q[0]
		q = append(q[1:], n.Children...)

		if n.Name != "" {
			logs.Debug(fmt.Sprintf(insertRef, moduleName, quoteEscape(n.Purl())))
			_, err := db.Exec("insert or ignore into reference (module_name, purl) values (?, ?)", moduleName, n.Purl())
			if err != nil {
				logs.Warn(err)
			}
		}

	}

}

// quoteEscape 转义单引号
func quoteEscape(src string) (out string) {
	out = strings.ReplaceAll(src, `'`, "\\'")
	return out
}
