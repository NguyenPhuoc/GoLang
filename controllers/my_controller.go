package controllers

import (
	"GoLang/config/configdb"
	"GoLang/config/environment"
	"GoLang/libraries/aes256"
	"GoLang/libraries/constants"
	"GoLang/libraries/logtype"
	"GoLang/libraries/sessionmng"
	"GoLang/libraries/util"
	"GoLang/models"
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"
	"github.com/kataras/iris"
	"github.com/kataras/iris/mvc"
	"github.com/kataras/iris/sessions"
	"github.com/mileusna/useragent"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

type MyController struct {
	mvc.BaseController
	Session       *sessions.Session
	DB            *gorm.DB
	RedisInfo     *redis.Client
	RedisConfig   *redis.Client
	Database      map[string]*gorm.DB
	Redises       map[string]*redis.Client
	MongoDB       *mongo.Client
	DataResponse  iris.Map
	User          models.HfUser
	IsAPI         IsAPI
	CheckTime     time.Time
	InsertedID    interface{}
	IsEncrypt     IsEncrypt
	IsMetaData    IsMetaData
	UserAgent     ua.UserAgent
	UserVersion   int
	ServerVersion int
}
type IsAPI bool
type IsEncrypt bool
type IsMetaData bool

type formValue func(string) string

func (c *MyController) BeforeActivation(b mvc.BeforeActivation) {
	b.Dependencies().Add(func(ctx iris.Context) formValue { return ctx.FormValue })
}

func (c *MyController) BeginRequest(ctx iris.Context) {
	c.CheckTime = time.Now()

	c.SetDB(0)

	UserAgent := ctx.GetHeader("user-agent")
	c.UserAgent = ua.Parse(UserAgent)
	c.UserVersion = util.ToInt(ctx.GetHeader("user-version"))
	platform := ctx.GetHeader("platform")
	if platform == "IOS" {
		c.UserAgent.OS = ua.IOS
	} else {
		c.UserAgent.OS = ua.Android
	}

	serverVersion := 0
	if c.UserAgent.OS == ua.IOS {
		serverVersion = util.ToInt(c.User.GetConfig("server_version_ios").Value)
	} else if c.UserAgent.OS == ua.Android {
		serverVersion = util.ToInt(c.User.GetConfig("server_version_android").Value)
	} else {
		serverVersion = util.ToInt(c.User.GetConfig("server_version_android").Value)
	}
	c.ServerVersion = serverVersion

	//Lưu thêm thông tin
	UserAgent += `[user-version:` + util.ToString(c.UserVersion) + `, platform:` + platform + `, server_version:` + util.ToString(serverVersion) + `, OS:` + c.UserAgent.OS + `]`


	//log request
	cooVal := ctx.GetCookie(sessionmng.COOKIE_NAME)

	collection := c.MongoDB.Database(configdb.LOG_MONGODB).Collection(logtype.HF_USER_REQUEST_API)
	res, err := collection.InsertOne(context.Background(), iris.Map{
		"server_id": nil, "user_id": nil, "full_name": nil, "path": ctx.GetCurrentRoute().Path(), "param": ctx.FormValues(), "response": nil, "user_agent": UserAgent, "cookie": cooVal, "duration": nil, "created_date": c.CheckTime,
	})
	if err != nil {
		fmt.Println("=========ERROR-InsertOne==========")
		fmt.Println(err)
	}
	c.InsertedID = res.InsertedID

	c.User = c.isAuth(ctx)

	//check ccu
	go c.Redises[configdb.CCU_REDIS].Set("1minute:"+c.User.UserId, true, time.Minute)
	go c.Redises[configdb.CCU_REDIS].Set("5minute:"+c.User.UserId, true, time.Minute*5)
	go c.Redises[configdb.CCU_REDIS].Set("15minute:"+c.User.UserId, true, time.Minute*15)

	go c.Redises[configdb.CCU_REDIS].Set("1request:"+util.UUID(), true, time.Minute)
	go c.Redises[configdb.CCU_REDIS].Set("5request:"+util.UUID(), true, time.Minute*5)
	go c.Redises[configdb.CCU_REDIS].Set("15request:"+util.UUID(), true, time.Minute*15)
}

func (c *MyController) EndRequest(ctx iris.Context) {
	if len(c.DataResponse) != 0 {
		env := environment.LoadEnvironmentLocal()
		if c.IsEncrypt && env.ENVIRONMENT != environment.LOCAL {
			_, _ = ctx.Text(aes256.Encrypt(util.JsonEndCode(c.DataResponse), constants.AES256_PASSPHRASE))
		} else {
			_, _ = ctx.JSON(c.DataResponse)
		}

		Cookie := ctx.GetHeader("Cookie")
		fmt.Println(util.TimeToDateTime(time.Now()), time.Now().Sub(c.CheckTime), ctx.GetCurrentRoute().Path(), c.User.UserName, Cookie)

		//if ctx.GetCurrentRoute().Path() == "/user/token" {
		//	uTokenLog := models.HfUserTokenLog{UserId: c.User.UserId, Data: util.JsonEndCode(c.DataResponse)}
		//	//go c.DB.Create(&uTokenLog)
		//	go func() {
		//		collection := c.MongoDB.Database(configdb.LOG_MONGODB).Collection(uTokenLog.TableName())
		//		res, _ := collection.InsertOne(context.Background(), iris.Map{
		//			"UserId": c.User.UserId, "Data": util.JsonEndCode(c.DataResponse), "CreatedDate": time.Now(),
		//		})
		//		res.InsertedID
		//	}()
		//}
	}

	param, _ := json.Marshal(ctx.FormValues())
	response, _ := json.Marshal(c.DataResponse)

	if !c.IsAPI {
		fmt.Println(util.TimeToDateTime(time.Now()), fmt.Sprintf("%s ?= %s", ctx.String(), string(param)))
		if c.IsMetaData {
			fmt.Println(util.TimeToDateTime(time.Now()), `{"IsMetaData":true}`)
		} else {
			fmt.Println(util.TimeToDateTime(time.Now()), string(response))
		}
		fmt.Println("-------------------------------")
		fmt.Println("-------------------------------")
	}

	//update log
	if c.IsMetaData {
		response = []byte(`{"IsMetaData":true}`)
	}
	res := util.JsonDecodeMap(util.JsonEndCode(c.DataResponse))
	collection := c.MongoDB.Database(configdb.LOG_MONGODB).Collection(logtype.HF_USER_REQUEST_API)
	_, err := collection.UpdateOne(context.Background(), iris.Map{"_id": c.InsertedID},
		iris.Map{"$set": iris.Map{"server_id": c.User.ServerId, "user_id": c.User.UserId, "full_name": c.User.FullName, "response": res, "duration": time.Now().Sub(c.CheckTime), "duration_string": time.Now().Sub(c.CheckTime).String()}},
	)
	if err != nil {
		fmt.Println("=========ERROR-UpdateOne==========")
		fmt.Println(err)
	}
}

func (c *MyController) isAuth(ctx iris.Context) models.HfUser {
	switch ctx.GetCurrentRoute().Path() {
	case "/demo/login", "/demo/signup", "/login/mobile", "/notlogin":
		c.IsAPI = true
	}

	if !c.IsAPI {
		//fmt.Println("Before.Session.Get", time.Now().Sub(c.CheckTime))
		signedRequest := c.Session.Get(sessionmng.KEY_SIGNED_REQUEST)
		//fmt.Println("After.Session.Get ", time.Now().Sub(c.CheckTime))
		dataSigned := util.InterfaceToMap(signedRequest)
		userId := util.ToString(dataSigned["user_id"])
		serverId := uint(util.ToInt(dataSigned["server_id"]))
		access := util.ToString(dataSigned["access"])

		if signedRequest == nil || userId == "" {
			ctx.Redirect("/notlogin")
			return c.User
		}
		UserAgent := ctx.GetHeader("User-Agent")
		Cookie := ctx.GetHeader("Cookie")
		fmt.Println(util.TimeToDateTime(time.Now()), util.JsonEndCode(signedRequest), Cookie)
		fmt.Println(util.TimeToDateTime(time.Now()), UserAgent)
		fmt.Println(util.TimeToDateTime(time.Now()), ctx.String(), util.JsonEndCode(ctx.FormValues()))
		fmt.Println()

		go c.Session.Lifetime.Shift(sessionmng.TIME_EXPIRES)
		cooVal := ctx.GetCookie(sessionmng.COOKIE_NAME)
		if cooVal != "" {
			ctx.SetCookieKV(sessionmng.COOKIE_NAME, cooVal, iris.CookieExpires(sessionmng.TIME_EXPIRES))
		}
		go c.Session.Set(sessionmng.KEY_SIGNED_REQUEST, signedRequest)

		//isAuth
		c.SetDB(serverId)
		user := c.User
		user.UserId = userId
		user = user.Get()

		if user.Access.String != access {
			ctx.Redirect("/notlogin")
			return c.User
		}

		go c.RedisInfo.Expire(userId, sessionmng.TIME_EXPIRES)

		return user

	}

	return c.User
}

func (c *MyController) validToken(ctx iris.Context) bool {
	token := ctx.FormValue("token")
	sign := ctx.FormValue("sign")

	valid, option := util.ValidToken(token, sign, c.Session)
	if !valid {
		c.DataResponse = iris.Map{
			"code":   -1,
			"msg":    "Invalid token",
			"option": option,
		}

		bug := c.Redises[configdb.BUG_REDIS]
		param, _ := json.Marshal(ctx.FormValues())
		signed_request := c.Session.Get(sessionmng.KEY_SIGNED_REQUEST)
		t := time.Now()
		key := fmt.Sprintf("%d-%02d-%02d:%02d-%02d-%02d:%d-%s", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), token)
		bug.HSet(key, "sign", sign)
		bug.HSet(key, "option", util.JsonEndCode(option))
		bug.HSet(key, "values", ctx.String())
		bug.HSet(key, "param", string(param))
		bug.HSet(key, "signed_request", util.JsonEndCode(signed_request))
	}

	return valid
}

func (c *MyController) SetDB(serverId uint, users ...models.HfUser) models.HfUser {

	//dbName := dbconnect.GetDBServer(serverId)
	dbName := configdb.MAIN_DB//db mặc định

	c.DB = c.Database[dbName]
	c.RedisInfo = c.Redises[configdb.INFO_REDIS]
	c.RedisConfig = c.Redises[configdb.CONFIG_REDIS]
	//c.MongoDB = c.MongoDB

	//user global
	c.User.DB = c.DB
	c.User.RedisInfo = c.RedisInfo
	c.User.RedisConfig = c.RedisConfig
	c.User.MongoDB = c.MongoDB

	if serverId != 0 {
		sv := models.HfServer{}
		dbName = sv.GetDBName(serverId, c.User)

		//Cập nhật lại db cho đúng
		c.DB = c.Database[dbName]
		c.User.DB = c.DB
	}

	// User param
	user := models.HfUser{}
	if len(users) > 0 {
		user = users[0]
	}

	user.DB = c.DB
	user.RedisInfo = c.RedisInfo
	user.RedisConfig = c.RedisConfig
	user.MongoDB = c.MongoDB

	return user
}
