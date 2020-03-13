package main

import (
	"GoLang/config/environment"
	"GoLang/controllers"
	"GoLang/libraries/dbconnect"
	"GoLang/libraries/sessionmng"
	"github.com/kataras/iris"
	"github.com/kataras/iris/mvc"
)

func main() {
	env := environment.LoadEnvironment()
	db := dbconnect.Init()
	sessMng := sessionmng.Init()

	app := iris.New()
	app.RegisterView(iris.HTML("../../public_html", ".html").Reload(true))
	app.StaticWeb("/", "../../public_html")

	index := mvc.New(app.Party("/"))
	index.Register(db.Databases, db.Redises, db.MongoDB, sessMng.Sessions.Start)
	index.Handle(new(controllers.IndexController))

	index.Clone(app.Party("/user")).Handle(new(controllers.UserController))
	index.Clone(app.Party("/inbox")).Handle(new(controllers.InboxController))
	index.Clone(app.Party("/arena")).Handle(new(controllers.ArenaController))
	index.Clone(app.Party("/pet")).Handle(new(controllers.PetController))
	index.Clone(app.Party("/summon")).Handle(new(controllers.SummonController))
	index.Clone(app.Party("/mission")).Handle(new(controllers.MissionController))
	index.Clone(app.Party("/newbie")).Handle(new(controllers.NewbieController))
	index.Clone(app.Party("/rollup")).Handle(new(controllers.RollupController))
	index.Clone(app.Party("/guardian")).Handle(new(controllers.GuardianController))
	index.Clone(app.Party("/cashshop")).Handle(new(controllers.CashshopController))
	index.Clone(app.Party("/event")).Handle(new(controllers.EventController))
	index.Clone(app.Party("/giftcode")).Handle(new(controllers.GiftCodeController))

	index.Clone(app.Party("/test")).Handle(new(controllers.TestController))

	var isAPI controllers.IsAPI = true
	var isEncrypt controllers.IsEncrypt = false
	var isMetaData controllers.IsMetaData = false
	API := mvc.New(app.Party("/api"))
	API.Register(db.Databases, db.Redises, db.MongoDB, sessMng.Sessions.Start, isAPI, isEncrypt, isMetaData)
	API.Handle(new(controllers.ApiController))
	API.Clone(app.Party("/config")).Handle(new(controllers.ConfigController))
	API.Clone(app.Party("/arena-test")).Handle(new(controllers.ArenaTestController))
	API.Clone(app.Party("/cron")).Handle(new(controllers.CronController))

	_ = app.Run(iris.Addr(env.Addr))
}
