package origin

import (
	"encoding/json"
	"os"
	"util/args"
	"util/logs"

	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func (o *BaseOrigin) LoadJsonOrigin(filepath string) {
	if jsonText, err := os.ReadFile(filepath); err != nil {
		logs.Error(err)
	} else {
		data := []VulnInfo{}
		err := json.Unmarshal(jsonText, &data)
		if err != nil {
			logs.Error(err)
		}
		o.LoadDataOrigin(data...)
	}
}

func (o *BaseOrigin) LoadMysqlOrigin(cfg args.OriginConfig) {
	o.LoadSqlOrigin(mysql.Open(cfg.Dsn), cfg)
}

func (o *BaseOrigin) LoadSqliteOrigin(cfg args.OriginConfig) {
	o.LoadSqlOrigin(sqlite.Open(cfg.Dsn), cfg)
}

func (o *BaseOrigin) LoadSqlOrigin(dialector gorm.Dialector, cfg args.OriginConfig) {
	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		logs.Error(err)
		return
	}
	data := []VulnInfo{}
	db.Table(cfg.Table).Find(&data)
	o.LoadDataOrigin(data...)
}
