package dbconnect

import (
	"GoLang/config/configdb"
	"GoLang/config/environment"
	"GoLang/libraries/util"
	"context"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"
	"github.com/kataras/iris"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"sync"
	"time"
)

type dbconnect struct {
	Databases map[string]*gorm.DB
	Redises   map[string]*redis.Client
	MongoDB   *mongo.Client
}

var conn *dbconnect
var once sync.Once

func Init() *dbconnect {
	once.Do(func() {
		conn = connect()
	})
	return conn
}

func connect() *dbconnect {
	confdb := configdb.LoadConfig()

	Databases := make(map[string]*gorm.DB)
	Redises := make(map[string]*redis.Client)
	//Todo DB
	for name, conf := range confdb.DB {
		strConn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local", conf.Username, conf.Password, conf.Host, conf.Port, conf.DBName, conf.CharSet)
		db, err := gorm.Open(conf.DBDriver, strConn)
		if err != nil {
			fmt.Println(err)
		} else {
			var waitTimeout time.Duration
			_ = db.DB().QueryRow("SELECT @@wait_timeout").Scan(&waitTimeout)

			db.DB().SetConnMaxLifetime(time.Second * waitTimeout)
			db.DB().SetMaxIdleConns(conf.MaxIdle)
			db.DB().SetMaxOpenConns(conf.MaxOpen)
			//fmt.Println("DB:", name, util.JsonEndCode(conf), waitTimeout*time.Second)

			Databases[name] = db
		}
	}

	//Todo Redis
	for name, conf := range confdb.Redis {
		client := redis.NewClient(&redis.Options{
			Network:  conf.Network,
			Addr:     conf.Addr,
			Password: conf.Password,
			DB:       util.ToInt(conf.Database),
		})
		_, err := client.Ping().Result()
		if err != nil {
			fmt.Println(err)
		} else {
			//fmt.Println("Red:",name, util.JsonEndCode(conf))

			Redises[name] = client
		}
	}

	//Todo MongoDB
	//mongodb://192.168.88.253:27017
	clientOptions := options.Client().ApplyURI(confdb.MongoConnect)

	env := environment.LoadEnvironmentLocal()
	if env.ENVIRONMENT == environment.LOCAL {
		clientOptions = options.Client().ApplyURI("mongodb://192.168.88.253:27017/log?retryWrites=true")
	}

	// Connect to MongoDB
	clientMongoDB, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		fmt.Println(err)
	}
	mongoPing := clientMongoDB.Ping(context.TODO(), readpref.Primary())
	if mongoPing != nil {
		fmt.Println(mongoPing)
	}

	iris.RegisterOnInterrupt(func() {
		for _, db := range Databases {
			_ = db.Close()
		}
		for _, db := range Redises {
			_ = db.Close()
		}
		_ = clientMongoDB.Disconnect(context.Background())
	})

	return &dbconnect{Databases: Databases, Redises: Redises, MongoDB: clientMongoDB}
}

func GetDBServer(serverId uint) string {

	config := map[uint]string{
		0: configdb.MAIN_DB,
		1: configdb.MAIN_DB,
		2: configdb.MAIN_DB,
		3: configdb.MAIN_DB,
		4: configdb.MAIN_DB,
		5: configdb.MAIN_DB,
	}

	if dbName, ok := config[serverId]; ok {
		return dbName
	}

	return config[0]
}
