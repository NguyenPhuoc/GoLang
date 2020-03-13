package configdb

import (
	"GoLang/config/environment"
	"github.com/tkanos/gonfig"
)

const (
	MAIN_DB string = "main"

	INFO_REDIS    string = "info"
	SESSION_REDIS string = "session"
	CONFIG_REDIS  string = "config"
	BUG_REDIS     string = "bug"
	CCU_REDIS     string = "ccu"
	LOG_MONGODB   string = "log"
)

type configdb struct {
	DB           map[string]config
	Redis        map[string]configRedis
	MongoConnect string
}

type config struct {
	Host     string
	Port     int
	Username string
	Password string
	DBName   string

	DBDriver string
	DBPrefix string

	CharSet string

	MaxIdle int
	MaxOpen int
}

type configRedis struct {
	Network     string
	Addr        string
	Password    string
	Database    string
	MaxIdle     int
	MaxActive   int
	IdleTimeout int
	Prefix      string
}

func LoadConfig() configdb {
	env := environment.LoadEnvironment()
	conf := configdb{}
	var err error

	if env.ENVIRONMENT == environment.LOCAL || env.ENVIRONMENT == environment.DEV {
		err = gonfig.GetConf("config/configdb/config.development.json", &conf)
	} else if env.ENVIRONMENT == environment.PRODUCTION {
		err = gonfig.GetConf("config/configdb/config.production.json", &conf)
	} else {
		err = gonfig.GetConf("config/configdb/config.development.json", &conf)
	}
	if err != nil {
		panic(err)
	}
	return conf
}
