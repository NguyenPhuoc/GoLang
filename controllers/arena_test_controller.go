package controllers

import (
	"GoLang/libraries/constants"
	"GoLang/libraries/logtype"
	"GoLang/libraries/util"
	"GoLang/models"
	"github.com/kataras/iris"
	"math"
	"sync"
	"time"
)

type ArenaTestController struct {
	MyController
}

// /arena-test/token
func (c *ArenaTestController) GetToken(ctx iris.Context) {
	token := util.GetToken(c.Session)

	c.DataResponse = iris.Map{
		"token": token,
	}
}

// /arena-test/token
func (c *ArenaTestController) PostToken(form formValue, ctx iris.Context) {
	if c.validToken(ctx) {
		c.DataResponse = iris.Map{"token": "Yesssssss"}
	} else {
		c.DataResponse = iris.Map{"token": "Noooooooo"}
	}
}

// /arena-test/winboss
func (c *ArenaTestController) PostWinboss(form formValue, ctx iris.Context) {
	if c.validToken(ctx) {
	//if true {
		user := c.User
		user.UserId = form("user_id")
		user.UserName = form("user_name")
		count := 0
		c.DB.Where("user_id = ?", user.UserId).First(&user).Count(&count)
		if count == 0 {
			user.ServerId = 0
			user.UserName = form("user_name")
			user.FullName = form("user_name")
			user.Password = "1"
			user.LastIncreasePower = util.Time()
			c.DB.Save(&user)
		}

		node := uint(util.ParseInt(form("node")))
		mapId := uint(util.ParseInt(form("map_id")))
		mapDiff := uint(util.ParseInt(form("map_diff")))
		win := uint8(util.ParseInt(form("win")))
		data := form("data")
		signHash := util.ValidHash(form, mapDiff, mapId, node, win)

		var wg sync.WaitGroup
		if signHash && node != 0 && mapId != 0 && mapDiff != 0 {
			wg.Add(5)

			//Todo Node hiện tại => Lấy Gift
			campCurrentNode := models.HfCampaign{}
			go func() {
				campCurrentNode = campCurrentNode.Find(mapDiff, mapId, node, user)
				wg.Done()
			}()

			//Todo Node User đang ở
			ucn := models.HfUserCampaignNode{}
			go func() {
				ucn = ucn.Get(user)
				wg.Done()
			}()

			//Todo Node cuối của map
			campLastNode := models.HfCampaign{}
			go func() {
				campLastNode = campLastNode.FindLastMap(mapDiff, mapId, user)
				wg.Done()
			}()

			//Todo Map cuối của Diff
			campLastMap := models.HfCampaign{}
			go func() {
				campLastMap = campLastMap.FindLastDiff(mapDiff, user)
				wg.Done()
			}()

			//Todo lấy Max Diff
			campMaxDiff := models.HfCampaign{}
			go func() {
				campMaxDiff = campMaxDiff.FindMaxDiff(user)
				wg.Done()
			}()
			wg.Wait()

			//fmt.Println("ucn",ucn.Node, ucn.MapId, ucn.MapDiff);
			//fmt.Println("campLastNode",campLastNode.Node, campLastNode.MapId, campLastNode.MapDiff);
			//fmt.Println("campLastMap",campLastMap.Node, campLastMap.MapId, campLastMap.MapDiff);
			//fmt.Println("campMaxDiff",campMaxDiff.Node, campMaxDiff.MapId, campMaxDiff.MapDiff);

			//Todo set nhảy Node nếu đang đánh thắng ở node cao nhất của user
			if ucn.Node == node && ucn.MapId == mapId && ucn.MapDiff == mapDiff && win == 1 {

				if ucn.Node == campLastMap.Node && ucn.MapId == campLastMap.MapId { //Todo set nhảy MapDiff nếu ở node, map, diff cuối
					if ucn.MapDiff < campMaxDiff.MapDiff { // kiểm tra max
						ucn.MapDiff += 1;
						ucn.MapId = 1
						ucn.Node = 1
					}
				} else if ucn.Node == campLastNode.Node { //Todo set nhảy Map nếu đang ở node cuối
					ucn.MapId += 1
					ucn.Node = 1
				} else {
					ucn.Node += 1
				}

				//Todo Nhảy Node Thám Hiểm
				uExplore := models.HfUserExplore{}
				uExplore, _, _ = uExplore.Get(user)

				uExplore.Node = ucn.Node
				uExplore.MapId = ucn.MapId
				//set lại time để nhảy node và cũng đã được cập nhật quantity node cũ trước ở trong uExplore.Get(user) rồi nên sẽ không bị tính lại
				uExplore.StartTimes = time.Now()

				wg.Add(2)
				go func() { //Todo lưu nhảy node
					c.DB.Save(&ucn)
					wg.Done()
				}()


				go uExplore.Save(user, &wg)
				//go func() { //Todo Node Thám hiểm
				//	c.DB.Save(&uExplore)
				//	wg.Done()
				//}()
				wg.Wait()
			}

			wg.Add(1)
			//Todo Lưu Log
			ucl := models.HfUserCampaignLog{UserId: user.UserId, Node: node, MapId: mapId, MapDiff:mapDiff, Win: win, Data: data}
			go func() {
				c.DB.Save(&ucl)
				wg.Done()
			}()

			//Todo Lấy quà
			gifts := map[string]interface{}{}
			if win == 1 {
				wg.Add(1)
				gifts = util.JsonDecodeMap(campCurrentNode.Gift)
				go user.UpdateGifts(gifts, logtype.GIFT_WIN_BOSS, 0, &wg)
			}
			wg.Wait()

			c.DataResponse = iris.Map{"code": 1, "data": iris.Map{"gift": gifts, "map_diff": ucn.MapDiff, "map_id": ucn.MapId, "node": ucn.Node}}
		} else {
			c.DataResponse = iris.Map{"code": -1}
		}
	}
}

// /arena-test/info/pvp
func (c *ArenaTestController) PostInfoPvp(form formValue) {
	user := c.User
	user.UserId = form("user_id")
	user.UserName = form("user_name")
	count := 0
	c.DB.Where("user_id = ?", user.UserId).First(&user).Count(&count)
	if count == 0 {
		user.ServerId = 0
		user.UserName = form("user_name")
		user.FullName = form("user_name")
		user.Password = "1"
		user.LastIncreasePower = util.Time()
		c.DB.Save(&user)
	}

	uArena := models.HfUserArenaPvp{}
	uArena = uArena.Get(user)
	if uArena.LastWin != -1 {
		isWin := uArena.LastWin == 1 //1: thắng, 0: thua
		uArena.EndgameWinLose(isWin, "", user)
	}

	myRank, tops := uArena.GetRankAndTop(c.DB)

	c.DataResponse = iris.Map{
		"code": 1,
		"data": iris.Map{
			"info":         uArena.GetMap(),
			"price_ticket": uArena.GetPriceTicket(),
			"rank":         myRank,
			"top":          util.JsonDecodeArray(tops),
		},
	}
}

// /arena-test/endgame/pvp
func (c *ArenaTestController) PostEndgamePvp(form formValue, ctx iris.Context) {
	user := c.User
	user.UserId = form("user_id")
	data := form("data")

	uArena := models.HfUserArenaPvp{}
	uArena = uArena.Get(user)
	uArena.Level = util.ParseInt(form("level"))
	uArena.PowerPoint = util.ParseInt(form("power_point"))

	isWin := uArena.LastWin == 1 //1: thắng, 0: thua

	if uArena.LastWin == -1 {
		c.DataResponse = iris.Map{"code": -1, "msg": "Invalid"}
		return
	}

	elo, gift := uArena.EndgameWinLose(isWin, data, user)

	c.DataResponse = iris.Map{"code": 1, "msg": "Success", "data": iris.Map{"elo_pvp": uArena.Elo, "delta_elo_pvp": elo, constants.GIFT: gift}}

}

// /arena-test/explore
func (c *ArenaTestController) PostExplore(form formValue, ctx iris.Context) {
	user := c.User
	user.UserId = form("user_id")
	user.UserName = form("user_name")
	count := 0
	c.DB.Where("user_id = ?", user.UserId).First(&user).Count(&count)
	if count == 0 {
		user.ServerId = 0
		user.UserName = form("user_name")
		user.FullName = form("user_name")
		user.Password = "1"
		user.LastIncreasePower = util.Time()
		c.DB.Save(&user)
	}

	uExplore := models.HfUserExplore{}

	timeNow := util.Time()

	uExplore, gifts, _ := uExplore.Get(user)
	nextTime := 5 - ((timeNow.Unix() - uExplore.StartTimes.Unix()) % 5)
	quantityBox := uExplore.GetQuantityBox()

	c.DataResponse = iris.Map{
		"code": 1,
		"data": iris.Map{
			"node":         uExplore.Node,
			"map_id":       uExplore.MapId,
			"gift":         gifts,
			"quantity_box": quantityBox,
			"next_time":    nextTime,
			"total_time":   math.Min(3600*23,float64(time.Now().Unix() - uExplore.StartTimes.Unix())),
		},
	}
}

// /arena-test/set/setting
func (c *ArenaTestController) PostSetSetting(form formValue, ctx iris.Context) {
	user := c.User

	user.UserId = form("user_id")
	user.UserName = form("user_name")
	count := 0
	c.DB.Where("user_id = ?", user.UserId).First(&user).Count(&count)
	if count == 0 {
		user.ServerId = 0
		user.UserName = form("user_name")
		user.FullName = form("user_name")
		user.Password = "1"
		user.LastIncreasePower = util.Time()
		c.DB.Save(&user)
	}

	setting := form("setting")

	userSetting := models.HfUserSetting{}
	userSetting.Get(user)

	userSetting.Setting = setting
	c.DB.Save(&userSetting)

	c.DataResponse = iris.Map{"code": 1, "msg": "Success"}
}
