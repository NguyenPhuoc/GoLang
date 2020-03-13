package controllers

import (
	"GoLang/config/cashshop"
	"GoLang/config/configdb"
	"GoLang/config/kulconfig"
	"GoLang/libraries/constants"
	"GoLang/libraries/sessionmng"
	"GoLang/libraries/util"
	"GoLang/models"
	"bytes"
	"database/sql"
	"github.com/go-sql-driver/mysql"
	"github.com/jinzhu/now"
	"github.com/kataras/iris"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"sync"
	"time"
)

type ApiController struct {
	MyController
}

// /api/kul/allserver
func (c *ApiController) PostKulAllserver(form formValue, ctx iris.Context) {
	dataLogin := util.JsonDecodeMap(form("data_login"))
	c.SetDB(0)

	res := util.InterfaceToMap(dataLogin)
	num, _ := strconv.ParseFloat(util.ToString(res["UserID"]), 64)
	partnerId := int64(num)

	sv := models.HfServer{}
	svIds := sv.GetDBs(c.User)

	results := []struct {
		ServerId  uint
		LastLogin time.Time
	}{}

	if partnerId != 0 {
		for _, id := range svIds {
			c.SetDB(id)

			res := []struct {
				ServerId  uint
				LastLogin time.Time
			}{}
			c.DB.Raw("SELECT server_id, last_login FROM hf_user WHERE partner_id = ? AND last_login IS NOT NULL ORDER BY last_login DESC LIMIT 10;", partnerId).Scan(&res)

			for _, r := range res {
				results = append(results, r)
			}
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].LastLogin.Unix() > results[j].LastLogin.Unix()
	})

	c.SetDB(0)
	allServer := sv.GetAll(c.User)
	allReturn := []iris.Map{}
	for _, val := range allServer {
		if val.ServerId != 999999 {
			allReturn = append(allReturn, iris.Map{"server_id": val.ServerId, "name": val.Name.String, "new": val.New})
		}
	}

	lastReturn := []iris.Map{}
	for _, val := range results {
		lastReturn = append(lastReturn, iris.Map{"server_id": val.ServerId})
	}

	if c.UserVersion > c.ServerVersion {
		allReturn = []iris.Map{{"server_id": 999999, "name": "S1", "new": 0}}
		lastReturn = []iris.Map{}
	}

	c.DataResponse = iris.Map{"code": 1, "data": iris.Map{
		"all_server":  allReturn,
		"last_server": lastReturn,
	}}
}

// /api/kul/login
func (c *ApiController) PostKulLogin(form formValue, ctx iris.Context) {
	dataLogin := util.JsonDecodeMap(form("data_login"))
	serverId := uint(util.ToInt(form("server_id")))
	c.SetDB(serverId)

	//if util.ToInt(dataLogin["e"]) != 0 {
	//	c.DataResponse = iris.Map{"code": -1, "msg": "Login fail e!=0"}
	//	return
	//}
	//
	//r := util.InterfaceToMap(dataLogin["r"])
	r := util.InterfaceToMap(dataLogin)
	accessToken := util.ToString(r["AccessToken"])

	kul := kulconfig.Load(accessToken)

	client := &http.Client{}

	data := url.Values{}
	data.Set("AppKey", kul.AppKey)
	data.Add("AccessToken", kul.AccessToken)
	data.Add("Time", kul.Time)
	data.Add("Sign", kul.Sign)

	req, err := http.NewRequest("POST", kul.API, bytes.NewBufferString(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	if err != nil {
		c.DataResponse = iris.Map{"code": -1, "msg": err.Error()}
		log.Println(err)
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		c.DataResponse = iris.Map{"code": -1, "msg": err.Error()}
		log.Println(err)
		return
	}

	f, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.DataResponse = iris.Map{"code": -1, "msg": err.Error()}
		log.Println(err)
		return
	}
	_ = resp.Body.Close()
	if err != nil {
		c.DataResponse = iris.Map{"code": -1, "msg": err.Error()}
		log.Println(err)
		return
	}
	res := util.JsonDecodeMap(string(f))

	if util.ToInt(res["e"]) != 0 {
		c.DataResponse = iris.Map{"code": -1, "msg": "Login server fail e!=0", "data": res}
		return
	}

	res = util.InterfaceToMap(res["r"])

	user := models.HfUser{}
	num, _ := strconv.ParseFloat(util.ToString(res["UserID"]), 64)
	user.PartnerId = int64(num)
	user.ServerId = serverId

	count := 0
	c.DB.Where("partner_id = ? and server_id = ?", user.PartnerId, user.ServerId).First(&user).Count(&count)
	if count == 0 {
		user.DB = c.DB
		user.UserId = util.UUID()
		user.PartnerId = int64(num)
		user.ServerId = serverId
		user.UserName = util.ToString(res["UserName"])
		//user.FullName = util.ToString(res["DisplayName"])
		user.FullName = util.ToString(res["UserName"])
		user.Password = "1"
		user.SignUp()

		go c.Redises[configdb.CCU_REDIS].Set("nru:"+user.UserId, true, now.EndOfDay().Sub(time.Now()))
	}
	go c.Redises[configdb.CCU_REDIS].Set("log:"+user.UserId, true, now.EndOfDay().Sub(time.Now()))

	c.Session.Lifetime.Shift(sessionmng.TIME_EXPIRES)
	cooVal := ctx.GetCookie(sessionmng.COOKIE_NAME)
	if cooVal != "" {
		ctx.SetCookieKV(sessionmng.COOKIE_NAME, cooVal, iris.CookieExpires(sessionmng.TIME_EXPIRES))
	}

	user.Access = sql.NullString{String: util.UUID(), Valid: true}
	user = c.SetDB(serverId, user)

	//Để lưu log
	c.User.UserId = user.UserId
	c.User.FullName = user.FullName

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

// /api/kul/getcharacter
func (c *ApiController) GetKulGetcharacter(form formValue, ctx iris.Context) {

	num, _ := strconv.ParseFloat(util.ToString(form("KulUserID")), 64)
	KulUserID := int64(num)
	ServerID := uint(util.ToInt(form("ServerID")))
	c.SetDB(ServerID)

	user := models.HfUser{}
	user.PartnerId = KulUserID
	user.ServerId = ServerID
	count := 0

	c.DB.Where("partner_id = ? and server_id = ?", user.PartnerId, user.ServerId).First(&user).Count(&count)
	if count == 0 {
		c.DataResponse = iris.Map{"e": -100, "r": "Role not Exists"}
		return
	}

	//Để lưu log
	c.User.UserId = user.UserId
	c.User.FullName = user.FullName

	user = c.SetDB(ServerID, user) //set server db cho user
	user = user.Get()

	//Lấy thông tin gói trưởng thành
	uCashshop := models.HfUserCashshop{}
	uCashshop.Get(user)
	firstPackage := map[string]uint8{}
	util.JsonDecodeObject(uCashshop.FirstPackage, &firstPackage)

	CardFoever := firstPackage["p10"]

	//{"e":0,"r":{"RoleName":"DaiKaSanCa","RoleID":100105,"Level":27,"Gold":99999,"CardFoever":1} }
	c.DataResponse = iris.Map{"e": 0, "r": iris.Map{"RoleName": user.UserName, "RoleID": user.UserId, "Level": user.GetLevel(), "Gold": user.Gold, "Gem": user.Gem, "CardFoever": CardFoever}}
}

// /api/kul/generateorderid
func (c *ApiController) PostKulGenerateorderid(form formValue, ctx iris.Context) {

	KulOrderID := form("KulOrderID")
	RoleID := form("RoleID")
	ServerID := uint(util.ToInt(form("ServerID")))
	PackageID := form("PackageID")
	Time := form("Time")
	Sign := form("Sign")

	//Để lưu log
	c.User.UserId = RoleID

	checkSign, checkMsg := kulconfig.CheckSign(Sign, KulOrderID, RoleID, ServerID, PackageID, Time)
	if !checkSign {
		c.DataResponse = iris.Map{"e": -102, "r": checkMsg}
		return
	}

	csConfig := cashshop.Config()
	if _, ok := csConfig[PackageID]; !ok {
		c.DataResponse = iris.Map{"e": -108, "r": "PackageID not Exists"}
		return
	}

	//Set db để check user
	c.SetDB(ServerID)
	user := models.HfUser{}
	countUser := 0

	c.DB.Where("user_id = ? and server_id = ?", RoleID, ServerID).First(&user).Count(&countUser)
	if countUser == 0 {
		c.DataResponse = iris.Map{"e": -100, "r": "RoleID not Exists"}
		return
	}

	//Set db để làm check transaction
	c.SetDB(0)
	uTran := models.HfUserTransaction{PartnerOrderId: sql.NullString{String: KulOrderID, Valid: true}}

	countKulOrder := 0
	c.DB.Where(uTran).First(&uTran).Count(&countKulOrder)
	if countKulOrder != 0 {
		c.DataResponse = iris.Map{"e": -101, "r": "Duplicate KulOrderID"}
		return
	}

	//Để lưu log
	c.User.UserId = user.UserId
	c.User.FullName = user.FullName

	uTran.OrderId = util.UUID()
	uTran.PartnerOrderId = sql.NullString{String: KulOrderID, Valid: true}
	uTran.UserId = RoleID
	uTran.ServerId = ServerID
	uTran.PackageId = PackageID
	uTran.TimeOrder = time.Now()
	uTran.Type = constants.PAYMENT_INWEB

	c.DB.Save(&uTran)

	c.DataResponse = iris.Map{"e": 0, "r": iris.Map{"GameOrderID": uTran.OrderId}}
}

// /api/kul/topup
func (c *ApiController) PostKulTopup(form formValue, ctx iris.Context) {

	GameOrderID := form("GameOrderID")
	KulOrderID := form("KulOrderID")
	RoleID := form("RoleID")
	ServerID := uint(util.ToInt(form("ServerID")))
	PackageID := form("PackageID")
	Time := form("Time")
	Sign := form("Sign")

	RedirectStatus := util.ToInt(form("RedirectStatus"))
	if ServerID == 999999 && RedirectStatus == 0 {

		client := &http.Client{}

		data := url.Values{}
		data.Set("GameOrderID", form("GameOrderID"))
		data.Add("KulOrderID", form("KulOrderID"))
		data.Add("RoleID", form("RoleID"))
		data.Add("ServerID", form("ServerID"))
		data.Add("PackageID", form("PackageID"))
		data.Add("Time", form("Time"))
		data.Add("Sign", form("Sign"))
		data.Add("RedirectStatus", "1")

		req, err := http.NewRequest("POST", "https://pokidemo-api.cala.games/api/kul/topup", bytes.NewBufferString(data.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
		if err != nil {
			c.DataResponse = iris.Map{"code": -1, "msg": err.Error()}
			log.Println(err)
			return
		}
		resp, err := client.Do(req)
		if err != nil {
			c.DataResponse = iris.Map{"code": -1, "msg": err.Error()}
			log.Println(err)
			return
		}

		f, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			c.DataResponse = iris.Map{"code": -1, "msg": err.Error()}
			log.Println(err)
			return
		}
		_ = resp.Body.Close()
		if err != nil {
			c.DataResponse = iris.Map{"code": -1, "msg": err.Error()}
			log.Println(err)
			return
		}
		c.DataResponse= util.JsonDecodeMap(string(f))
		return
	}

	//Để lưu log
	c.User.UserId = RoleID

	checkSign, checkMsg := kulconfig.CheckSign(Sign, KulOrderID, RoleID, ServerID, PackageID, Time, GameOrderID)
	if !checkSign {
		c.DataResponse = iris.Map{"e": -102, "r": checkMsg}
		return
	}

	//Check GameOrderID
	uTran := models.HfUserTransaction{OrderId: GameOrderID}

	countGameOrder := 0
	c.DB.Where(uTran).First(&uTran).Count(&countGameOrder)
	if countGameOrder == 0 {
		c.DataResponse = iris.Map{"e": -101, "r": "GameOrderID not Exists"}
		return
	}

	if uTran.IsTopup != 0 {
		c.DataResponse = iris.Map{"e": -107, "r": "Order has Topup"}
		return
	}

	if uTran.PackageId != PackageID {
		c.DataResponse = iris.Map{"e": -103, "r": "PackageID Invalid"}
		return
	}

	if uTran.PartnerOrderId.Valid && uTran.PartnerOrderId.String != KulOrderID {
		c.DataResponse = iris.Map{"e": -104, "r": "KulOrderID Invalid"}
		return
	}
	if !uTran.PartnerOrderId.Valid {

		tempTran := models.HfUserTransaction{PartnerOrderId: sql.NullString{String: KulOrderID, Valid: true}}

		countKulOrder := 0
		c.DB.Where(tempTran).First(&tempTran).Count(&countKulOrder)
		if countKulOrder != 0 {
			c.DataResponse = iris.Map{"e": -105, "r": "Duplicate KulOrderID"}
			return
		}

		uTran.PartnerOrderId = sql.NullString{String: KulOrderID, Valid: true}
	}

	csConfig := cashshop.Config()
	if _, ok := csConfig[PackageID]; !ok {
		c.DataResponse = iris.Map{"e": -108, "r": "PackageID not Exists"}
		return
	}

	RoleID = uTran.UserId
	ServerID = uTran.ServerId

	//Để lưu log
	c.User.UserId = RoleID

	//Set db để check user và nạp
	c.SetDB(ServerID)
	user := models.HfUser{}
	countUser := 0

	c.DB.Where("user_id = ? and server_id = ?", RoleID, ServerID).First(&user).Count(&countUser)
	if countUser == 0 {
		c.DataResponse = iris.Map{"e": -100, "r": "RoleID not Exists"}
		return
	}

	//Để lưu log
	c.User.UserId = user.UserId
	c.User.FullName = user.FullName

	user = c.SetDB(ServerID, user) //set server db cho user
	user = user.Get()

	user.Payment(PackageID, uTran.OrderId, uTran.Type == constants.PAYMENT_INWEB)

	//set lại server để lưu trạng thái order
	c.SetDB(0)
	uTran.IsTopup = 1
	uTran.TimeTopup = mysql.NullTime{Time: time.Now(), Valid: true}
	c.DB.Save(&uTran)

	c.DataResponse = iris.Map{"e": 0, "r": iris.Map{"GameOrderID": uTran.OrderId}}
}
