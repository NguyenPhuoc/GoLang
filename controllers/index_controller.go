package controllers

import (
	"GoLang/config/environment"
	"GoLang/libraries/constants"
	"GoLang/libraries/sessionmng"
	"GoLang/libraries/util"
	"GoLang/models"
	"database/sql"
	"fmt"
	"github.com/jinzhu/now"
	"github.com/kataras/iris"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type IndexController struct {
	MyController
}

func (c *IndexController) Get(ctx iris.Context) {
	_ = ctx.View("index.html")
}

func (c *IndexController) GetDemo(ctx iris.Context) {
	_ = ctx.View("demo.html")
}

func (c *IndexController) GetUtilsIndex(ctx iris.Context) {
	_ = ctx.View("utils/index.html")
}

func (c *IndexController) GetDemoLogin(ctx iris.Context) {
	ctx.ViewData("error_login", c.Session.GetFlash("error_login"))
	_ = ctx.View("login.html")
}

func (c *IndexController) GetNotlogin(ctx iris.Context) {
	c.DataResponse = iris.Map{"code": -1, "msg": "not logged in"}
}

func (c *IndexController) GetDemoLogout(ctx iris.Context) {
	c.Session.Destroy()
	ctx.Redirect("/demo/login")
}

func (c *IndexController) PostDemoSignup(form formValue, ctx iris.Context) {
	signUpType := form("type")
	user_name := form("user_name")
	password := form("password")

	count := 0
	user := models.HfUser{UserName: user_name}
	c.DB.Where(user).First(&user).Count(&count)

	if count >= 1 {
		c.DataResponse = iris.Map{"code": -1, "msg": "User name invalid"}
		return
	}
	if signUpType == "check" && count == 0 {
		c.DataResponse = iris.Map{"code": 1, "msg": "User name valid"}
		return
	} else {

		env := environment.LoadEnvironmentLocal()

		user.ServerId = 0
		if env.ENVIRONMENT == environment.PRODUCTION {
			user.ServerId = 1
		}
		user.UserId = util.UUID()
		user.FullName = user_name
		user.Password = password

		user = c.SetDB(user.ServerId, user)

		user.SignUp()

	}
	c.DataResponse = iris.Map{"code": 1, "msg": "Sign up success"}
}

func (c *IndexController) PostDemoLogin(form formValue, ctx iris.Context) {
	user_name, password, server_id := form("user_name"), form("password"), uint(0)
	c.SetDB(server_id)

	user := models.HfUser{}
	count := 0
	c.DB.Where("server_id = ? and user_name = ? and password = ?", server_id, user_name, password).First(&user).Count(&count)
	if count == 0 {
		c.Session.SetFlash("error_login", true)

		ctx.Redirect("/demo/login")
	} else {

		c.Session.Lifetime.Shift(sessionmng.TIME_EXPIRES)
		cooVal := ctx.GetCookie(sessionmng.COOKIE_NAME)
		if cooVal != "" {
			ctx.SetCookieKV(sessionmng.COOKIE_NAME, cooVal, iris.CookieExpires(sessionmng.TIME_EXPIRES))
		}
		//c.Session.Set(sessionmng.KEY_SIGNED_REQUEST, iris.Map{
		//	"user_id":   user.UserId,
		//	"user_name": user.UserName,
		//	"server_id": user.ServerId,
		//})

		user.Access = sql.NullString{String: util.UUID(), Valid: true}
		user.DB = c.DB
		user.RedisInfo = c.RedisInfo

		var wg sync.WaitGroup
		wg.Add(1)
		go user.UpdateKey("Access", user.Access, &wg)
		wg.Wait()

		c.Session.Set(sessionmng.KEY_SIGNED_REQUEST, iris.Map{
			"user_id":   user.UserId,
			"user_name": user.UserName,
			"server_id": user.ServerId,
			"access":    user.Access.String,
		})

		ctx.Redirect("/")
	}
}

// /login/mobile
func (c *IndexController) PostLoginMobile(form formValue, ctx iris.Context) {
	user_name, password, server_id := form("user_name"), form("password"), uint(0)
	c.SetDB(server_id)

	user := models.HfUser{}
	count := 0
	//c.DB.Where("server_id = ? and user_name = ? and password = ?", server_id, user_name, password).First(&user).Count(&count)
	c.DB.Where("user_name = ? and password = ?", user_name, password).First(&user).Count(&count)
	if count == 0 {
		fmt.Println("Login fail", util.JsonEndCode(user))
		c.DataResponse = iris.Map{"code": -1, "msg": "Login fail"}
	} else {

		c.Session.Lifetime.Shift(sessionmng.TIME_EXPIRES)
		cooVal := ctx.GetCookie(sessionmng.COOKIE_NAME)
		if cooVal != "" {
			ctx.SetCookieKV(sessionmng.COOKIE_NAME, cooVal, iris.CookieExpires(sessionmng.TIME_EXPIRES))
		}

		user.Access = sql.NullString{String: util.UUID(), Valid: true}
		user.DB = c.DB
		user.RedisInfo = c.RedisInfo

		var wg sync.WaitGroup
		wg.Add(1)
		go user.UpdateKey("Access", user.Access, &wg)
		wg.Wait()

		c.Session.Set(sessionmng.KEY_SIGNED_REQUEST, iris.Map{
			"user_id":   user.UserId,
			"user_name": user.UserName,
			"server_id": user.ServerId,
			"access":    user.Access.String,
		})

		c.DataResponse = iris.Map{"code": 1, "msg": "Login success"}
	}
}

func (c *IndexController) GetInit(ctx iris.Context) {
	c.IsMetaData = true
	user := c.User
	var wg sync.WaitGroup

	wg.Add(1)
	pet, pets := models.HfUserPet{}, []iris.Map{}
	go func() {
		pets = pet.GetPetEquip(user)
		wg.Done()
	}()

	wg.Add(1)
	userSetting := models.HfUserSetting{}
	go func() {
		userSetting.Get(user)
		wg.Done()
	}()

	wg.Add(1)
	ucNode := models.HfUserCampaignNode{}
	go func() {
		ucNode.Get(user)
		wg.Done()
	}()

	wg.Add(1)
	stones := make(map[string]uint)
	go func() {
		stones = user.GetStones()
		wg.Done()
	}()

	wg.Add(1)
	uEquip, equips := models.HfUserEquip{}, []iris.Map{}
	go func() {
		equips = uEquip.GetEquipMap(user)
		wg.Done()
	}()

	wg.Add(1)
	uPieceGeneral, sPieceGeneral := models.HfUserPiece{}, ""
	go func() {
		sPieceGeneral = uPieceGeneral.GetPieces(user)
		wg.Done()
	}()

	wg.Add(1)
	uEvent := models.HfUserEvent{}
	go uEvent.Init(user, &wg)

	wg.Add(1)
	uGuardian, guardians := models.HfUserGuardian{}, []iris.Map{}
	go func() {
		guardians = uGuardian.GetMaps(user)
		wg.Done()
	}()

	wg.Wait()

	//Update last login
	go user.CheckLastLogin()

	info := iris.Map{
		"user_id":   user.UserId,
		"server_id": user.ServerId,
		"full_name": user.FullName,
		"gold":      user.Gold,
		"gem":       user.Gem,
		"power":     user.Power,
		"exp":       user.Exp,
	}

	campaign_node := iris.Map{
		"node":   ucNode.Node,
		"map_id": ucNode.MapId,
	}

	startArenaPvp, _ := now.Parse("18:00")
	endArenaPvp, _ := now.Parse("22:00")
	timeServer := map[string]interface{}{
		"server": time.Now().Unix(),
		"arena_pvp": []interface{}{
			map[string]int64{"start": startArenaPvp.Unix(), "end": endArenaPvp.Unix()},
		},
	}
	data := iris.Map{
		"info":          info,
		"step_newbie":   user.NewbieStep,
		"server_time":   util.TimeUnix(),
		"time":          timeServer,
		"pets":          pets,
		"setting":       util.JsonDecodeMap(userSetting.Setting),
		"campaign_node": campaign_node,
		"stones":        stones,
		"equips":        equips,
		"piece_general": util.JsonDecodeArray(sPieceGeneral),
		"guardians":     guardians,
	}

	c.DataResponse = iris.Map{
		"code": 1,
		"data": data,
	}
}

func (c *IndexController) GetInitV2(ctx iris.Context) {
	c.IsEncrypt = true
	user := c.User

	//Todo rule báo bảo trì từ init
	user = c.SetDB(0, user)
	whiteList := models.HfWhitelist{}
	whiteLists := whiteList.GetAll(user)

	maintenance, _ := strconv.ParseBool(whiteLists[0].Value)
	if maintenance {
		checkPass := false
		for i := 1; i < len(whiteLists); i++ {
			wl := whiteLists[i]
			switch wl.Type {
			case "ip":
			case "user_name":
				if user.UserName == wl.Value {
					checkPass = true
					break
				}
			case "full_name":
				if user.FullName == wl.Value && user.ServerId == wl.ServerId {
					checkPass = true
					break
				}
			case "partner_id":
				if util.ToString(user.PartnerId) == wl.Value {
					checkPass = true
					break
				}
			}
		}

		if !checkPass {
			c.DataResponse = iris.Map{"code": -2, "msg": "maintenance"}
			return
		}
	}
	user = c.SetDB(user.ServerId, user)

	//Todo check có bị ban hay không
	if user.Banned != 0 {
		uBan := models.HfUserBanned{}
		uBan, check := uBan.Find(user)

		if uBan.FreedomDay.Unix() <= time.Now().Unix() || !check {
			uBan.UnBan(user)
		} else {
			c.DataResponse = iris.Map{"code": -3, "data": iris.Map{"reason": uBan.Reason, "freedom_day": util.TimeHHiissddmmYYYY(uBan.FreedomDay)}}
			return
		}
	}

	var wg sync.WaitGroup
	c.IsMetaData = true

	wg.Add(1)
	pet, pets := models.HfUserPet{}, []iris.Map{}
	go func() {
		pets = pet.GetPetEquip(user)
		wg.Done()
	}()

	wg.Add(1)
	userSetting := models.HfUserSetting{}
	go func() {
		userSetting.Get(user)
		wg.Done()
	}()

	wg.Add(1)
	ucNode := models.HfUserCampaignNode{}
	campMaxNodeMap := models.HfCampaign{}
	go func() {
		ucNode.Get(user)
		campMaxNodeMap = campMaxNodeMap.FindMaxNodeMap(user)
		wg.Done()
	}()

	wg.Add(1)
	stones := make(map[string]uint)
	go func() {
		stones = user.GetStones()
		wg.Done()
	}()

	wg.Add(1)
	uEquip, equips := models.HfUserEquip{}, []iris.Map{}
	go func() {
		equips = uEquip.GetEquipMap(user)
		wg.Done()
	}()

	wg.Add(1)
	uPieceGeneral, sPieceGeneral := models.HfUserPiece{}, ""
	go func() {
		sPieceGeneral = uPieceGeneral.GetPieces(user)
		wg.Done()
	}()

	wg.Add(1)
	uEvent := models.HfUserEvent{}
	go uEvent.Init(user, &wg)

	wg.Add(1)
	uGuardian, guardians := models.HfUserGuardian{}, []iris.Map{}
	go func() {
		guardians = uGuardian.GetMaps(user)
		wg.Done()
	}()

	wg.Add(1)
	uGuaInfo := models.HfUserGuardianInfo{}
	go func() {
		uGuaInfo.Get(user)
		wg.Done()
	}()

	wg.Add(1)
	ueRename, renameGemFee := models.HfUserEvent{}, 0
	go func() {
		ueRename = ueRename.GetRename(user)
		renameGemFee = ueRename.GetRenameFee(user)
		wg.Done()
	}()

	wg.Add(1)
	gemSoul := 0
	go func() {
		gemSoul = pet.GetGemSoul(user)
		wg.Done()
	}()
	wg.Wait()

	//Update last login
	go user.CheckLastLogin()

	info := iris.Map{
		"user_id":    user.UserId,
		"server_id":  user.ServerId,
		"full_name":  user.FullName,
		"gold":       user.Gold,
		"gem":        user.Gem,
		"power":      user.Power,
		"exp":        user.Exp,
		"rename_fee": renameGemFee,
		"gem_soul":   gemSoul,
		"avatar_id":  user.AvatarId,
	}

	campaign_node := iris.Map{
		"node":    ucNode.Node,
		"map_id":  ucNode.MapId,
		"is_last": campMaxNodeMap.Node == ucNode.Node && campMaxNodeMap.MapId == ucNode.MapId,
	}

	startArenaPvp := now.MustParse("12:00")
	endArenaPvp := now.MustParse("22:00")
	timeServer := map[string]interface{}{
		"server": time.Now().Unix(),
		"arena_pvp": []interface{}{
			map[string]int64{"start": startArenaPvp.Unix(), "end": endArenaPvp.Unix()},
		},
	}

	materials := iris.Map{
		constants.FRUIT_GUARDIAN:  uGuaInfo.Fruit,
		constants.FLOWER_GUARDIAN: uGuaInfo.Flower,
	}

	data := iris.Map{
		"info":          info,
		"step_newbie":   user.NewbieStep,
		"server_time":   util.TimeUnix(),
		"time":          timeServer,
		"pets":          pets,
		"setting":       util.JsonDecodeMap(userSetting.Setting),
		"campaign_node": campaign_node,
		"stones":        stones,
		"equips":        equips,
		"piece_general": util.JsonDecodeArray(sPieceGeneral),
		"guardians":     guardians,
		"materials":     materials,
	}

	c.DataResponse = iris.Map{
		"code": 1,
		"data": data,
	}
}

// /readfile
func (c *IndexController) PostReadfile(form formValue, ctx iris.Context) {
	if c.validToken(ctx) {
		GOPATH := os.Getenv("GOPATH") // /data/www/pokiwarh5
		path := fmt.Sprintf("%s/public_html/%s", GOPATH, form("file"))

		if strings.Contains(path, "..") {
			c.DataResponse = iris.Map{"code": -1, "msg": "..."}
			return
		}

		data, err := ioutil.ReadFile(path)
		if err != nil || strings.Contains(path, "..") {
			c.DataResponse = iris.Map{"code": -1, "msg": err.Error()}
			return
		}
		c.DataResponse = iris.Map{"code": 1, "data": string(data)}
	}
}
