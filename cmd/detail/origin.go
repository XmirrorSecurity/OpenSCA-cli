package detail

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/xmirrorsecurity/opensca-cli/v3/cmd/config"
	"github.com/xmirrorsecurity/opensca-cli/v3/opensca/logs"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	_origin *BaseOrigin
	_once   = sync.Once{}
)

type VulnInfo struct {
	*Vuln
	Vendor   string `json:"vendor" gorm:"column:vendor"`
	Product  string `json:"product" gorm:"column:product"`
	Version  string `json:"version" gorm:"column:version"`
	Language string `json:"language" gorm:"column:language"`
}

type BaseOrigin struct {
	// origin data
	// map[language]map[component_name][]VulnInfo
	data  map[string]map[string][]VulnInfo
	idSet map[string]bool
}

func NewBaseOrigin() *BaseOrigin {
	return &BaseOrigin{
		data:  map[string]map[string][]VulnInfo{},
		idSet: map[string]bool{},
	}
}

func (o *BaseOrigin) LoadDataOrigin(data ...VulnInfo) {
	if o == nil {
		return
	}
	for _, info := range data {
		if info.Vuln == nil {
			continue
		}
		if o.idSet[info.Id] {
			continue
		}
		o.idSet[info.Id] = true
		name := strings.ToLower(info.Product)
		language := strings.ToLower(info.Language)
		if _, ok := o.data[language]; !ok {
			o.data[language] = map[string][]VulnInfo{}
		}
		vulns := o.data[language]
		vulns[name] = append(vulns[name], info)
	}
}

func GetOrigin() *BaseOrigin {
	_once.Do(func() {
		_origin = NewBaseOrigin()
		c := config.Conf().Origin
		_origin.LoadJsonOrigin(c.Json)
		_origin.LoadMysqlOrigin(c.Mysql)
		_origin.LoadSqliteOrigin(c.Sqlite)
		logs.Info(fmt.Sprintf("load %d vulnerability", len(_origin.idSet)))
	})
	return _origin
}

func (o *BaseOrigin) LoadJsonOrigin(filepath string) {
	if filepath == "" {
		return
	}
	if jsonFile, err := os.Open(filepath); err != nil {
		logs.Error(err)
	} else {
		data := []VulnInfo{}
		err = json.NewDecoder(jsonFile).Decode(&data)
		if err != nil {
			logs.Error(err)
		}
		o.LoadDataOrigin(data...)
	}
}

func (o *BaseOrigin) LoadMysqlOrigin(cfg config.SqlOrigin) {
	o.LoadSqlOrigin(mysql.Open(cfg.Dsn), cfg)
}

func (o *BaseOrigin) LoadSqliteOrigin(cfg config.SqlOrigin) {
	o.LoadSqlOrigin(sqlite.Open(cfg.Dsn), cfg)
}

func (o *BaseOrigin) LoadSqlOrigin(dialector gorm.Dialector, cfg config.SqlOrigin) {
	if cfg.Dsn == "" {
		return
	}
	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.New(log.Default(), logger.Config{
			SlowThreshold: 1 * time.Second,
			LogLevel:      logger.Info,
		}),
	})
	if err != nil {
		logs.Error(err)
		return
	}
	data := []VulnInfo{}
	db.Table(cfg.Table).Find(&data)
	o.LoadDataOrigin(data...)
}
