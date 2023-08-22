package origin

import (
	"fmt"
	"sync"

	"github.com/xmirrorsecurity/opensca-cli/util/args"
	"github.com/xmirrorsecurity/opensca-cli/util/logs"
)

var (
	_origin *BaseOrigin
	_once   = sync.Once{}
)

func GetOrigin() *BaseOrigin {
	_once.Do(func() {
		_origin = NewBaseOrigin()
		if args.Config.DB != "" {
			_origin.LoadJsonOrigin(args.Config.DB)
		}
		for originType, config := range args.Config.Origin {
			switch originType {
			case "mysql":
				_origin.LoadMysqlOrigin(config)
			case "sqlite", "sqlite3":
				_origin.LoadSqliteOrigin(config)
			case "json":
				_origin.LoadJsonOrigin(config.Dsn)
			case "":
				// pass
			default:
				logs.Warn(fmt.Sprintf("not support origin type: %s", originType))
			}
		}
		logs.Info(fmt.Sprintf("load %d vulnerability", len(_origin.idSet)))
	})
	return _origin
}
