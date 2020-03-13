package controllers

import (
	"GoLang/config/cashshop"
	"GoLang/libraries/constants"
	"GoLang/libraries/logtype"
	"GoLang/libraries/util"
	"GoLang/models"
	"crypto/md5"
	"database/sql"
	"fmt"
	"github.com/kataras/iris"
	"github.com/mileusna/useragent"
	"io"
	"strings"
	"sync"
	"time"
)

type TestController struct {
	MyController
}

func (c *TestController) Get(ctx iris.Context) {

	a := 1.1
	b := 1.9
	abc := uint(a)
	abd := uint(b)

	//h := models.HfUserArenaPvpLog{UserId:"0000"}
	//c.DB.Create(&h)
	//c.DB.Create(&models.HfUserArenaPvpLog{UserId:"0000"})
	//fmt.Println(util.JsonEndCode(h))

	c.DataResponse = iris.Map{
		"code": 1,
		"data": iris.Map{
			"hello": "this's api test",
		},
		"abc": abc,
		"abd": abd,
	}
}

func (c *TestController) Post(form formValue, ctx iris.Context) {

	a := util.ParseInt(form("a"))
	jMap := util.JsonDecodeMap(form("a"))
	jArr := util.JsonDecodeArray(form("a"))

	token := form("token")
	hash := md5.New()
	_, _ = io.WriteString(hash, constants.HASH_TOKEN)
	_, _ = io.WriteString(hash, token)
	signHash := fmt.Sprintf("%x", hash.Sum(nil))

	c.DataResponse = iris.Map{
		"code": 1,
		"data": iris.Map{
			"hello": "this's api test",
		},
		"a":    a,
		"jMap": jMap,
		"jArr": jArr,
		"token": iris.Map{
			"signHash": signHash,
			"token":    token,
		},
	}
}

// /test/add/trangbi
func (c *TestController) GetAddTrangbi(form formValue, ctx iris.Context) {
	user := c.User
	ids := form("ids")
	quantity := uint16(util.ParseInt(form("quantity")))

	idsMap := strings.Split(ids, ",")
	for _, id := range idsMap {
		idInt := uint16(util.ParseInt(id))

		if idInt != 0 {
			ue := models.HfUserEquip{}
			ue.UserId = user.UserId
			ue.EquipId = idInt

			check := 0
			c.DB.Where(ue).First(&ue).Count(&check)
			if check == 0 {
				ue.Id = fmt.Sprintf("%s_trangbi_%d", user.UserId, idInt)
			}

			ue.Quantity += quantity
			c.DB.Save(&ue)

			fmt.Println(ue, check)
		}
	}

	c.DataResponse = iris.Map{
		"ids":      ids,
		"quantity": quantity,
	}
}

// /test/add/pet
func (c *TestController) GetAddPet(form formValue, ctx iris.Context) {
	user := c.User
	ids := form("ids")

	idsMap := strings.Split(ids, ",")
	for _, id := range idsMap {
		idInt := uint16(util.ParseInt(id))

		if idInt != 0 {
			up := models.HfUserPet{}
			up.UserId = user.UserId
			up.PetId = idInt

			check := 0
			c.DB.Where(up).First(&up).Count(&check)
			if check == 0 {
				up.Id = fmt.Sprintf("%s_pet_%d", user.UserId, idInt)
				up.Level = 1
			}
			up.Stat = 1

			c.DB.Save(&up)

			fmt.Println(up, check)
		}
	}

	c.DataResponse = iris.Map{
		"ids": ids,
	}
}

// /test/payment?p=1
func (c *TestController) GetPayment(form formValue, ctx iris.Context) {
	p := `p` + form("p")

	csConfig := cashshop.Config()
	if _, ok := csConfig[p]; !ok {
		c.DataResponse = iris.Map{"r": "package invalid"}
		return
	}

	c.User.Payment(p, "", true)
	c.DataResponse = iris.Map{"e": 0, "r": iris.Map{"Amount": 0, "OrderID": "", "Other": "", "Package": p, "Time": 0, "TimeSDKServer": 0}}
}

// /test/add/acc
func (c *TestController) GetAddAcc(form formValue, ctx iris.Context) {
	user_name := form("user_name")
	full := util.ParseInt(form("full"))

	user := models.HfUser{UserId: user_name}
	check := 0

	c.DB.Where(user).First(&user).Count(&check)
	if check == 0 {
		user.DB = c.DB
		user.ServerId = 0
		user.UserName = user_name
		user.FullName = user_name
		user.Password = "1"
		user.SignUp()
	}

	if full == 1 {
		user := models.HfUser{UserId: user_name}
		c.DB.Where(user).First(&user).Count(&check)

		user.Gold += 1000000
		user.Gem += 100000
		user.Power = 500
		c.DB.Save(&user)

		user.DB = c.DB
		var wg sync.WaitGroup
		wg.Add(1)
		gift := iris.Map{
			"stones": iris.Map{//'d','l','w','t','e','f','en','evo'
				"d": 10000,
				"l": 10000,
				"w": 10000,
				"t": 10000,
				"e": 10000,
				"f": 10000,
				"en": 10000,
				"evo": 10000,
			},
		}
		pet := models.HfPet{}
		pets := pet.GetAll(c.User)

		piece := map[uint16]int{}

		for _, val := range pets {
			piece[val.Id] = 1000
		}

		gift["piece"] = piece

		go user.UpdateGifts(gift, 0, 0, &wg)
		wg.Wait()

		list_equip := []uint16{101,102,103,104,105,106,107,108,109,110,111,112,113,114,115,116,117,118,119,120,201,202,203,204,205,206,207,208,209,210,211,212,213,214,215,216,217,218,219,220,301,302,303,304,305,306,307,308,309,310,311,312,313,314,315,316,317,318,319,320,401,402,403,404,405,406,407,408,409,410,411,412,413,414,415,416,417,418,419,420,501,502,503,504,505,506,507,508,509,510,511,512,513,514,515,516,517,518,519,520,601,602,603,604,605,606,607,608,609,610,611,612,613,614,615,616,617,618,619,620}
		var quantity uint16 = 10
		for _, idInt := range list_equip {

			ue := models.HfUserEquip{}
			ue.UserId = user.UserId
			ue.EquipId = idInt

			check := 0
			c.DB.Where(ue).First(&ue).Count(&check)
			if check == 0 {
				ue.Id = fmt.Sprintf("%s_trangbi_%d", user.UserId, idInt)
			}

			ue.Quantity += quantity
			c.DB.Save(&ue)
		}


	}

	c.DataResponse = iris.Map{
		"user":  user,
		"check": check,
	}
}


// /test/add/gift
func (c *TestController) GetAddGift(form formValue, ctx iris.Context) {
	user:= c.User

	var wg sync.WaitGroup
	wg.Add(1)
	gift := util.JsonDecodeMap(`{"gold":10,"stones":{"evo":10},"exp":10,"piece":{"3001":10,"3002":10},"equip":{"101":10,"102":10}}`)
	go user.UpdateGifts(gift, 0, 0, &wg)
	wg.Wait()
}

// /test/flush/cache
func (c *TestController) GetFlushCache() {
	user := c.User

	user.RedisInfo.Del(user.UserId)

	c.DataResponse = iris.Map{"code": 1, "msg": "Done"}

}

func (c *TestController) GetAbc() {
	user := c.User

	giftCf := map[int]interface{}{
		1: iris.Map{constants.GEM: 20000, constants.PIECE: map[int]int{3005: 50}},
		2: iris.Map{constants.GEM: 20000, constants.GOLD: 50000},
		3: iris.Map{constants.GEM: 20000, constants.PIECE: map[int]int{3020: 50}},
		4: iris.Map{constants.GEM: 25000, constants.GOLD: 75000},
		5: iris.Map{constants.GEM: 25000, constants.PIECE: map[int]int{3009: 50}},
		6: iris.Map{constants.GEM: 30000, constants.GOLD: 100000},
		7: iris.Map{constants.GEM: 50000, constants.PIECE: map[int]int{3013: 50}, constants.STONES: map[string]int{"f": 200, "e": 200, "t": 200, "w": 200, "l": 200, "d": 200}, constants.GOLD: 100000},
	}

	for i, _ := range giftCf {

		ui := models.HfUserInbox{}
		ui.Id = util.UUID()
		ui.ReceiverId = user.UserId
		ui.SenderType = constants.INBOX_SENDER_BY_EVENT
		ui.TypeLog = logtype.GIFT_EVENT_INBOX
		ui.EventId = 1
		ui.KindId = i
		ui.IsReceive = 0

		title := fmt.Sprintf("Quà mở server ngày %d", i)

		ui.Title = title
		ui.Gift = sql.NullString{Valid: true, String: util.JsonEndCode(giftCf[i])}
		ui.CreatedDate = time.Now()
		user.DB.Save(&ui)

	}


	c.DataResponse = iris.Map{"code": 1, "msg": "Done"}

}

// /test/os
func (c *TestController) AnyOs(value formValue) {

	osTest := ua.Parse(value("UserAgent"))

	c.DataResponse = iris.Map{"code": 1, "data": iris.Map{"UserAgent": c.UserAgent, "osTest": osTest}}

}