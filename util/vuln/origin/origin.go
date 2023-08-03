package origin

import (
	"fmt"
	"sync"
	"util/args"
	"util/logs"
)

var (
	_origin *BaseOrigin
	_once   = sync.Once{}
)

func GetOrigin() *BaseOrigin {
	_once.Do(func() {
		_origin = NewBaseOrigin()
		for originType, config := range args.Config.Origin {
			switch originType {
			case "mysql":
				_origin.LoadMysqlOrigin(config)
			case "sqlite", "sqlite3":
				_origin.LoadSqliteOrigin(config)
			case "":
				// pass
			default:
				logs.Warn(fmt.Sprintf("not support origin type: %s", originType))
			}
		}
	})
	return _origin
}
