package detail

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/xmirrorsecurity/opensca-cli/opensca/logs"
	"github.com/xmirrorsecurity/opensca-cli/util/args"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
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
		if info.Description != "" {
			info.DescriptionEn = ""
		}
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
