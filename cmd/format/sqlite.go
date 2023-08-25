package format

import (
	"database/sql"
	_ "embed"
	"os"
	"path/filepath"
	"strings"

	"github.com/xmirrorsecurity/opensca-cli/cmd/detail"
	"github.com/xmirrorsecurity/opensca-cli/opensca/logs"

	_ "github.com/glebarez/go-sqlite"
)

//go:embed sqlite_init.sql
var initSql string

func Sqlite(report Report, out string) {

	dbFile := out
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

	moduleName := filepath.Base(report.AppName)

	if report.DepDetailGraph != nil && report.DepDetailGraph.Name != "" {
		moduleName = report.DepDetailGraph.Name
	}

	logs.Debugf("---- sql report of %s", moduleName)
	insertFmt := "insert or ignore into component (name, version, vendor, language, purl) values ('%s','%s','%s','%s','%s');\n"
	insertRef := "insert or ignore into reference (module_name, purl) values ('%s','%s');\n"

	report.DepDetailGraph.ForEach(func(n *detail.DepDetailGraph) bool {
		if n.Name == "" {
			return true
		}
		logs.Debugf(insertFmt, quoteEscape(n.Name), quoteEscape(n.Version), quoteEscape(n.Vendor), quoteEscape(n.Language), quoteEscape(n.Purl()))
		_, err := db.Exec("insert or ignore into component (name, version, vendor, language, purl) values (?,?,?,?,?)", n.Name, n.Version, n.Vendor, n.Language, n.Purl())
		if err != nil {
			logs.Warn(err)
		}
		return true
	})

	report.DepDetailGraph.ForEach(func(n *detail.DepDetailGraph) bool {
		if n.Name == "" {
			return true
		}
		logs.Debugf(insertRef, moduleName, quoteEscape(n.Purl()))
		_, err := db.Exec("insert or ignore into reference (module_name, purl) values (?, ?)", moduleName, n.Purl())
		if err != nil {
			logs.Warn(err)
		}
		return true
	})

}

// quoteEscape 转义单引号
func quoteEscape(src string) (out string) {
	return strings.ReplaceAll(src, `'`, "\\'")
}
