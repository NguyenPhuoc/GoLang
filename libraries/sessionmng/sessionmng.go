package sessionmng

import (
	"GoLang/config/configdb"
	"github.com/kataras/iris"
	"github.com/kataras/iris/sessions"
	"github.com/kataras/iris/sessions/sessiondb/redis"
	"github.com/kataras/iris/sessions/sessiondb/redis/service"
	"sync"
	"time"
)

const (
	COOKIE_NAME        = "h5-api"
	TIME_EXPIRES       = time.Second * 86400
	KEY_SIGNED_REQUEST = "signed_request"
	KEY_TOKEN          = "token"
	EXPIRES_TOKEN      = 10
)

type sessionmng struct {
	Sessions *sessions.Sessions
}

var sess *sessionmng
var once sync.Once

func Init() *sessionmng {
	once.Do(func() {
		sess = connect()
	})
	return sess
}

func connect() *sessionmng {
	//db, err := badger.New("data/sessions")

	confdb := configdb.LoadConfig()

	confSession := confdb.Redis[configdb.SESSION_REDIS]

	config := service.DefaultConfig()
	config.Addr = confSession.Addr
	config.Database = confSession.Database

	db := redis.New(config)

	iris.RegisterOnInterrupt(func() {
		_ = db.Close()
	})

	sess := sessions.New(sessions.Config{Cookie: COOKIE_NAME, Expires: TIME_EXPIRES})
	sess.UseDatabase(db)

	//if err != nil {
	//	panic(err)
	//}

	return &sessionmng{Sessions: sess}
}
