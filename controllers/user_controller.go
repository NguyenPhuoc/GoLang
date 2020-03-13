package controllers

import (
	"GoLang/libraries/constants"
	"GoLang/libraries/logtype"
	"GoLang/libraries/util"
	"GoLang/models"
	"crypto/md5"
	"database/sql"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"github.com/kataras/iris"
	"io"
	"math"
	"sync"
	"time"
)

type UserController struct {
	MyController
}

func (c *UserController) GetToken(ctx iris.Context) {
	token := util.GetToken(c.Session)
	//hash := md5.New()
	//_, _ = io.WriteString(hash, constants.HASH_TOKEN)
	//_, _ = io.WriteString(hash, token)
	//signHash := fmt.Sprintf("%x", hash.Sum(nil))

	c.DataResponse = iris.Map{
		"token": token,
		//"sign": signHash,
	}
	//if c.User.UserName == "phuocnh" {
	//	user := models.HfUser{UserId:"traingau"}
	//	user = c.SetDB(1, user) //set server db cho user
	//	user = user.Get()
	//
	//	user.Payment("p9", "", true)
	//}
	//var wg sync.WaitGroup
	//wg.Add(2)
	//data := `{"list_turn":["e7b378db-401e-45a2-830b-3a442378ad1d_0","luz26","e7b378db-401e-45a2-830b-3a442378ad1d_0","luz26","e7b378db-401e-45a2-830b-3a442378ad1d_0"],"users":[{"data_init":{"mhp":67343,"mmp":300,"hp":67343,"mp":0,"ap":0},"data_damage":{"3518":[{"bonus_more_damage":0.0,"type_damage":"shoot","list_piece_in_turn":[0,0,3,6,5,0],"pet_id":3006,"star":0,"level":22,"enhance":0,"evolve":0}],"9231":[{"bonus_more_damage":0.0,"type_damage":"beat","list_piece_in_turn":[0,0,3,6,5,0],"pet_id":3006,"star":0,"level":22,"enhance":0,"evolve":0}],"7656":[{"bonus_more_damage":0.0,"type_damage":"beat","list_piece_in_turn":[0,0,3,6,5,0],"pet_id":3006,"star":0,"level":22,"enhance":0,"evolve":0}],"6382":[{"bonus_more_damage":0.0,"type_damage":"shoot","list_piece_in_turn":[0,0,3,6,5,0],"pet_id":3006,"star":0,"level":22,"enhance":0,"evolve":0}],"17719":[{"bonus_more_damage":0.0,"type_damage":"skill","list_piece_in_turn":[0,0,3,6,5,0],"pet_id":3006,"star":0,"level":22,"enhance":0,"evolve":0}],"10171":[{"bonus_more_damage":0.0,"type_damage":"skill","list_piece_in_turn":[0,0,3,6,5,0],"pet_id":3008,"star":0,"level":22,"enhance":0,"evolve":0}]},"main_pet":{"pet_id":3006,"star":0,"level":22,"enhance":0,"evolve":0},"support_pet":[{"pet_id":3008,"star":0,"level":22,"enhance":0,"evolve":0},{"pet_id":3019,"star":0,"level":22,"enhance":0,"evolve":0},{"pet_id":3020,"star":0,"level":23,"enhance":0,"evolve":0}],"name":"e7b378db-401e-45a2-830b-3a442378ad1d_0","auto_support":false},{"data_init":{"mhp":35552,"mmp":300,"hp":35552,"mp":0,"ap":0},"data_damage":{"1381":[{"bonus_more_damage":0.0,"type_damage":"beat","list_piece_in_turn":[0,0,3,6,5,0],"pet_id":1,"star":0,"level":30,"enhance":0,"evolve":1}],"526":[{"bonus_more_damage":0.0,"type_damage":"shoot","list_piece_in_turn":[0,0,3,6,5,0],"pet_id":1,"star":0,"level":30,"enhance":0,"evolve":1}],"880":[{"bonus_more_damage":0.0,"type_damage":"beat","list_piece_in_turn":[0,0,3,6,5,0],"pet_id":1,"star":0,"level":30,"enhance":0,"evolve":1}],"1391":[{"bonus_more_damage":0.0,"type_damage":"skill","list_piece_in_turn":[0,0,3,6,5,0],"pet_id":29,"star":0,"level":1,"enhance":0,"evolve":0}]},"main_pet":{"pet_id":1,"star":0,"level":30,"enhance":0,"evolve":1},"support_pet":[{"pet_id":29,"star":0,"level":1,"enhance":0,"evolve":0}],"name":"luz26"}],"swapA":3,"swapB":2,"countPieceA":34,"countPieceB":25,"duration":33.97474,"os":"Android"}`
	//go c.User.SaveLogMongo("abc", iris.Map{"user_id": c.User.UserId, "data": data, "created_date": time.Now()}, &wg)
	//go c.User.SaveLogMongo("abc", iris.Map{"user_id": c.User.UserId, "data": util.JsonDecodeMap(data), "created_date": time.Now()}, &wg)
	//wg.Wait()
}

func (c *UserController) PostToken(form formValue, ctx iris.Context) {
	token := "bc2bb38da780aef36555c2d3821ef204"
	sign := "b9e07fbe988b8d493757054bb068d1c0"
	hash := md5.New()
	_, _ = io.WriteString(hash, constants.HASH_TOKEN)
	_, _ = io.WriteString(hash, token)
	signHash := fmt.Sprintf("%x", hash.Sum(nil))

	fmt.Println(sign, signHash, sign == signHash)

	if c.validToken(ctx) {
		c.DataResponse = iris.Map{"token": "Yesssssss"}
	} else {
		c.DataResponse = iris.Map{"token": "Noooooooo"}
	}
}

// /user/set/setting
func (c *UserController) PostSetSetting(form formValue, ctx iris.Context) {
	if c.validToken(ctx) {
		setting := form("setting")

		userSetting := models.HfUserSetting{}
		userSetting.Get(c.User)

		userSetting.Setting = setting
		c.DB.Save(&userSetting)

		c.DataResponse = iris.Map{"code": 1, "msg": "Success"}
	}
}

// /user/winboss
func (c *UserController) PostWinboss(form formValue, ctx iris.Context) {
	if c.validToken(ctx) {
		user := c.User
		node := uint(util.ParseInt(form("node")))
		mapId := uint(util.ParseInt(form("map_id")))
		mapDiff := uint(util.ParseInt(form("map_diff")))
		win := uint8(util.ParseInt(form("win")))
		data := form("data")

		signHash := util.ValidHash(form, mapDiff, mapId, node, win)

		if !(signHash && node != 0 && mapId != 0 && mapDiff != 0) {
			c.DataResponse = iris.Map{"code": -1}
			return
		}

		var wg sync.WaitGroup
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
		gifts := map[string]interface{}{}
		if ucn.Node == node && ucn.MapId == mapId && ucn.MapDiff == mapDiff && win == 1 {
			checkGift := false

			if ucn.Node == campLastMap.Node && ucn.MapId == campLastMap.MapId { //Todo set nhảy MapDiff nếu ở node, map, diff cuối
				if ucn.MapDiff < campMaxDiff.MapDiff { // kiểm tra max
					ucn.MapDiff += 1;
					ucn.MapId = 1
					ucn.Node = 1

					checkGift = true
				}
			} else if ucn.Node == campLastNode.Node { //Todo set nhảy Map nếu đang ở node cuối
				ucn.MapId += 1
				ucn.Node = 1

				checkGift = true
			} else {
				ucn.Node += 1

				checkGift = true
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

			//Todo Lấy quà
			if checkGift {//kiểu check map cuối đã nhận chưa bằng cách có pass được chưa
				wg.Add(1)
				gifts = util.JsonDecodeMap(campCurrentNode.Gift)
				go user.UpdateGifts(gifts, logtype.GIFT_WIN_BOSS, 0, &wg)
			}

			wg.Wait()
		}

		wg.Add(1)
		//Todo Lưu Log
		ucl := models.HfUserCampaignLog{UserId: user.UserId, Node: node, MapId: mapId, MapDiff: mapDiff, Win: win, Data: data}
		//go func() {
		//	c.DB.Save(&ucl)
		//	wg.Done()
		//}()
		go user.SaveLogMongo(ucl.TableName(),iris.Map{"server_id": user.ServerId, "user_id": user.UserId, "node": node, "map_id": mapId, "map_diff": mapDiff, "win": win, "data": util.JsonDecodeMap(data), "created_time": time.Now()}, &wg)
		wg.Wait()

		c.DataResponse = iris.Map{"code": 1, "data": iris.Map{"gift": gifts, "map_diff": ucn.MapDiff, "map_id": ucn.MapId, "node": ucn.Node}}
	}
}

// /user/explore
func (c *UserController) GetExplore(ctx iris.Context) {
	user := c.User
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
			"total_time":   math.Min(3600*23, float64(time.Now().Unix()-uExplore.StartTimes.Unix())),
		},
	}
}

// /user/gift/explore
func (c *UserController) PostGiftExplore(ctx iris.Context) {
	if c.validToken(ctx) {
		user := c.User
		uExplore := models.HfUserExplore{}

		uExplore, gifts, giftsReturn := uExplore.Get(user)

		var wg sync.WaitGroup

		//Reset time và quà thám hiểm
		uExplore.StartTimes = time.Now()
		uExplore.Quantity = 0

		//Todo Kiểm tra quà Box để nhận
		quantityBox := uExplore.GetQuantityBox()
		giftBox := map[string]interface{}{}
		if quantityBox > 0 {
			giftBox = util.InterfaceToMap(uExplore.GetBox(user))

			//Cập nhật lại time quà box
			uExplore.StartBox = time.Now()

			wg.Add(1)
			go user.UpdateGifts(giftBox, logtype.GIFT_EXPLORE_BOX, 0, &wg)
			wg.Wait()
		}

		wg.Add(3)
		go uExplore.Save(user, &wg)
		go user.UpdateGifts(gifts, logtype.GIFT_EXPLORE, 0, &wg)
		//Todo Update nhiệm vụ
		go user.CompleteMissionDaily(models.QDL_CAMPAIGN, &wg)
		wg.Wait()

		giftsReturn = util.MergeGift(giftsReturn, giftBox)

		c.DataResponse = iris.Map{
			"code": 1,
			"data": iris.Map{
				"gift": giftsReturn,
			},
		}
	}
}

// /user/gift/box
// Todo không dùng nữa gộp chung vào PostGiftExplore
func (c *UserController) PostGiftBox(ctx iris.Context) {
	if c.validToken(ctx) {
		user := c.User
		uExplore := models.HfUserExplore{}

		uExplore.Get(user)
		if uExplore.GetQuantityBox() == 0 {
			c.DataResponse = iris.Map{"code": -1, "msg": "Box invalid"}
			return
		}

		gifts := uExplore.GetBox(user)
		if gifts == nil {
			c.DataResponse = iris.Map{"code": -1, "msg": "Gift invalid"}
			return
		}
		giftUpdate := util.InterfaceToMap(gifts)

		var wg sync.WaitGroup

		//reset time box
		uExplore.StartBox = time.Now()

		wg.Add(2)
		go uExplore.Save(user, &wg)

		go user.UpdateGifts(giftUpdate, logtype.GIFT_EXPLORE_BOX, 0, &wg)
		wg.Wait()

		c.DataResponse = iris.Map{
			"code": 1,
			"data": iris.Map{
				"gift": gifts,
			},
		}
	}
}

// /user/increase/power
func (c *UserController) GetIncreasePower(ctx iris.Context) {
	user := c.User

	timeNow := util.Time()
	var blockTime int64 = 300 //5 Phút
	var blockPower uint16 = 1 //1 Năng Lượng

	nextTime := blockTime - ((timeNow.Unix() - user.LastIncreasePower.Unix()) % blockTime)
	powerPlus := uint16((timeNow.Unix()-user.LastIncreasePower.Unix())/blockTime) * blockPower

	user.Power += powerPlus

	var maxPower uint16 = 500
	if powerPlus > 0 {
		if user.Power > maxPower {
			user.Power = maxPower
		}

		user.LastIncreasePower = timeNow
		c.DB.Save(&user)
	}

	c.DataResponse = iris.Map{
		"code": 1,
		"data": iris.Map{
			"power":     user.Power,
			"next_time": nextTime,
		},
	}
}

// /user/trace/error
func (c *UserController) PostTraceError(form formValue, ctx iris.Context) {
	//if c.validToken(ctx) {
	if true {
		user := c.User

		uTrError := models.HfUserTraceError{UserId: user.UserId, Data: form("data")}
		c.DB.Create(&uTrError)

		c.DataResponse = iris.Map{"code": 1, "msg": "Success"}
	}
}

// /user/tower/info
func (c *UserController) GetTowerInfo(ctx iris.Context) {
	user := c.User

	uTower := models.HfUserTower{}
	uTower.Get(user)
	uTower.IncreasePower(user)

	towerConfig := models.HfTowerConfig{}
	towerConfig, checkConfig := towerConfig.Find(uTower.Floor, c.DB)
	maxFloor := towerConfig.MaxFloor(c.DB)

	floor := iris.Map{}
	if checkConfig {
		floor = towerConfig.GetMap()
	}

	if uTower.CheckPower != 0 {
		var wg sync.WaitGroup
		powerFee := uTower.CheckPower

		uTower.CheckPower = 0
		wg.Add(1)
		go uTower.SetPower(powerFee, "", uTower.Floor, logtype.USE_GAME_TOWER, 0, user, &wg)
		wg.Wait()
	}

	myRank, tops := uTower.GetRankAndTop(c.DB)
	nextTime := -1
	if uTower.Power < uTower.MaxPower() {
		nextTime = uTower.BlockTime() - (int(time.Now().Unix()-uTower.UpdatePower.Unix()) % uTower.BlockTime())
	}

	c.DataResponse = iris.Map{
		"code": 1,
		"data": iris.Map{
			constants.POWER_TOWER: uTower.Power,
			"floor":               floor,
			"max_floor":           maxFloor,
			"rank":                myRank,
			"top":                 util.JsonDecodeArray(tops),
			"next_time":           nextTime,
			"max_time":            uTower.BlockTime(),
		},
	}
}

// /user/tower/info/v2
func (c *UserController) GetTowerInfoV2(ctx iris.Context) {
	user := c.User

	uTower := models.HfUserTower{}
	uTower.Get(user)
	uTower.IncreasePower(user)

	towerConfig := models.HfTowerConfig{}
	towerConfig, checkConfig := towerConfig.Find(uTower.Floor, c.DB)
	maxFloor := towerConfig.MaxFloor(c.DB)

	floorMap := iris.Map{}
	if checkConfig {
		floorMap = towerConfig.GetMap()
	}

	if uTower.CheckPower != 0 {
		var wg sync.WaitGroup
		powerFee := uTower.CheckPower

		uTower.CheckPower = 0
		wg.Add(1)
		go uTower.SetPower(powerFee, "", uTower.Floor, logtype.USE_GAME_TOWER, 0, user, &wg)
		wg.Wait()
	}

	myRank, tops := uTower.GetRankAndTop(c.DB)
	nextTime := -1
	if uTower.Power < uTower.MaxPower() {
		nextTime = uTower.BlockTime() - (int(time.Now().Unix()-uTower.UpdatePower.Unix()) % uTower.BlockTime())
	}

	c.IsEncrypt = true
	c.DataResponse = iris.Map{
		"code": 1,
		"data": iris.Map{
			constants.POWER_TOWER: uTower.Power,
			"floor":               floorMap,
			"max_floor":           maxFloor,
			"rank":                myRank,
			"top":                 util.JsonDecodeArray(tops),
			"next_time":           nextTime,
			"max_time":            uTower.BlockTime(),
		},
	}
}

// /user/buy/power
func (c *UserController) PostBuyPower(form formValue, ctx iris.Context) {
	if c.validToken(ctx) {
		user := c.User

		quantity := math.Max(float64(util.ParseInt(form("quantity"))), 1)

		uTower := models.HfUserTower{}
		uTower = uTower.Get(user)

		pricePower := uTower.GetPricePower(int(quantity))

		if user.Gem < uint(pricePower) {
			c.DataResponse = iris.Map{"code": -1, "msg": "no gem"}
			return
		}

		gemFee := pricePower * -1

		var wg sync.WaitGroup
		wg.Add(2)
		go uTower.SetPower(int(quantity), "", 0, logtype.BUY_POWER_TOWER, 0, user, &wg)
		go user.SetGem(gemFee, "", 0, logtype.BUY_POWER_TOWER, 0, &wg)
		wg.Wait()

		c.DataResponse = iris.Map{"code": 1, "msg": "Buy success", "data": iris.Map{constants.GIFT: iris.Map{constants.POWER_TOWER: uTower.Power}, constants.GEM: user.Gem}}
	}
}

// /user/tower/startgame
func (c *UserController) PostTowerStartgame(form formValue, ctx iris.Context) {
	if c.validToken(ctx) {
		user := c.User
		uTower := models.HfUserTower{}
		uTower = uTower.Get(user)

		if uTower.Power < 1 {
			c.DataResponse = iris.Map{"code": -1, "msg": "Not enough power"}
			return
		}

		uTower.CheckPower = -1
		c.DB.Save(&uTower)

		c.DataResponse = iris.Map{"code": 1, "msg": "Success"}
	}
}

// /user/tower/endgame
func (c *UserController) PostTowerEndgame(form formValue, ctx iris.Context) {
	if c.validToken(ctx) {
		user := c.User
		win := util.ParseInt(form("win"))
		floor := util.ParseInt(form("floor"))
		powerPoint := util.ParseInt(form("power_point"))
		level := util.ParseInt(form("level"))
		last_line_up := form("last_line_up")

		isWin := win == 1 //1: thắng, 0: thua

		uTower := models.HfUserTower{}
		uTower = uTower.Get(user)

		if !util.ValidHash(form, win, floor, level, powerPoint, last_line_up) {
			c.DataResponse = iris.Map{"code": -1, "msg": "Invalid hash"}
			return
		}

		if uTower.Floor != floor {
			c.DataResponse = iris.Map{"code": -1, "msg": "uTower.Floor !=  floor"}
			return
		}

		//Check năng lượng tháp
		//if uTower.CheckPower == 0 {
		//	c.DataResponse = iris.Map{"code": -1, "msg": "Invalid CheckPower"}
		//	return
		//}

		towerConfig := models.HfTowerConfig{}
		towerConfig, checkConfig := towerConfig.Find(uTower.Floor, c.DB)
		if !checkConfig {
			c.DataResponse = iris.Map{"code": 1, "msg": "Can't find floor config"}
			return
		}

		//Save log
		var wg sync.WaitGroup
		data := form("data")
		utl := models.HfUserTowerLog{UserId: user.UserId, Floor: uTower.Floor, IsWin: int8(win), Data: data}
		//go c.DB.Create(&utl)
		wg.Add(1)
		go user.SaveLogMongo(utl.TableName(), iris.Map{"server_id": user.ServerId, "user_id": user.UserId, "floor": uTower.Floor, "is_win": int8(win), "data": util.JsonDecodeMap(data), "created_date": time.Now()}, &wg)

		if isWin {
			wg.Add(2)

			uTower.CheckPower = 0
			uTower.PowerPoint = powerPoint
			uTower.Level = level
			uTower.Floor += 1
			uTower.LastLineUp = sql.NullString{String: last_line_up, Valid: true}

			go func() {
				c.DB.Save(&uTower)
				wg.Done()
			}()

			gifts := util.JsonDecodeMap(towerConfig.Gift)
			go user.UpdateGifts(gifts, logtype.GIFT_WIN_TOWER, 0, &wg)
			wg.Wait()
			c.DataResponse = iris.Map{"code": 1, "msg": "Success", "data": iris.Map{"gift": gifts}}
			return
		} else {
			powerFee := uTower.CheckPower

			uTower.CheckPower = 0
			wg.Add(1)
			go uTower.SetPower(powerFee, "", uTower.Floor, logtype.USE_GAME_TOWER, 0, user, &wg)
			wg.Wait()
		}

		c.DataResponse = iris.Map{"code": 1, "msg": "Success"}
	}
}

// /user/gift/level
func (c *UserController) PostGiftLevel(form formValue, ctx iris.Context) {
	if c.validToken(ctx) {
		user := c.User

		levelGift := user.LevelGift
		levelCurr := user.GetLevel()

		if levelGift >= levelCurr {
			c.DataResponse = iris.Map{"code": -1, "msg": "Hasn't gift"}
			return
		}

		lv := models.HfLevel{}
		levels := lv.GetAll(user)
		gifts := map[string]interface{}{}
		for _, lv := range levels {
			if lv.Level > levelGift && lv.Level <= levelCurr {
				giftLevel := util.JsonDecodeMap(lv.Gift)
				gifts = util.MergeGift(gifts, giftLevel)
			}
		}

		var wg sync.WaitGroup
		//Cập nhật quà
		wg.Add(1)
		go user.UpdateGifts(gifts, logtype.GIFT_LEVEL_UP, 0, &wg)
		wg.Wait()

		//Cập nhật trạng thái nhận quà
		user.LevelGift = levelCurr
		wg.Add(1)
		go user.UpdateKey("LevelGift", user.LevelGift, &wg)
		wg.Wait()

		c.DataResponse = iris.Map{"code": 1, "msg": "Success", "data": iris.Map{constants.GIFT: gifts}}
	}
}

// /user/view/profile
func (c *UserController) PostViewProfile(form formValue, ctx iris.Context) {
	if c.validToken(ctx) {

		userId := form("user_id")
		serverId := uint(util.ToInt(form("server_id")))
		typeInfo := form("type") //arena, arena_pvp, tower

		c.SetDB(serverId)

		u := models.HfUser{}
		count := 0
		c.DB.Where("user_id = ? and server_id = ?", userId, serverId).First(&u).Count(&count)
		if count == 0 {
			c.DataResponse = iris.Map{"code": -1, "msg": "Can't found user"}
			return
		}

		user := c.User
		user.UserId = userId
		user = user.Get()

		var wg sync.WaitGroup
		wg.Add(2)
		uArena := models.HfUserArena{}
		go func() {
			uArena.Get(user)
			wg.Done()
		}()

		uArenaPvp := models.HfUserArenaPvp{}
		go func() {
			uArenaPvp.Get(user)
			wg.Done()
		}()
		wg.Wait()

		lineUp := map[string]interface{}{}
		powerPoint := 0

		switch typeInfo {
		case "arena":
			lineUp = util.JsonDecodeMap(uArena.LastLineUp.String)
			powerPoint = uArena.PowerPoint
		case "arena_pvp":
			lineUp = util.JsonDecodeMap(uArenaPvp.LastLineUp.String)
			powerPoint = uArenaPvp.PowerPoint
		case "tower":
			uTower := models.HfUserTower{}
			uTower = uTower.Get(user)

			lineUp = util.JsonDecodeMap(uTower.LastLineUp.String)
			powerPoint = uTower.PowerPoint
		default:
			lineUp = util.JsonDecodeMap(uArena.LineUp.String)
			powerPoint = uArena.PowerPoint
		}

		data := iris.Map{
			"full_name":  user.FullName,
			"guild_name": "",
			"avatar_id":  user.AvatarId,
			"arena": iris.Map{
				"elo": uArena.Elo,
			},
			"line_up":     lineUp,
			"power_point": powerPoint,

			"arena_pvp": iris.Map{"elo": uArenaPvp.Elo},
			"exp":       user.Exp,
		}

		c.DataResponse = iris.Map{"code": 1, "data": data}
	}
}

// /user/rename
func (c *UserController) PostRename(form formValue, ctx iris.Context) {
	if c.validToken(ctx) {
		user := c.User
		fullName := form("full_name")

		if len([]rune(fullName)) < 4 || len([]rune(fullName)) > 16 {
			c.DataResponse = iris.Map{"code": -1, "msg": "Length Full name invalid"}
			return
		}

		u := models.HfUser{}
		count := 0
		c.DB.Where("full_name = ? and server_id = ?", fullName, user.ServerId).First(&u).Count(&count)
		if count != 0 {
			c.DataResponse = iris.Map{"code": -2, "msg": "Full name Duplicate"}
			return
		}
		var wg sync.WaitGroup

		ue := models.HfUserEvent{}
		ue = ue.GetRename(user)
		gemFee := ue.GetRenameFee(user)
		if gemFee > 0 {
			if user.Gem < uint(gemFee) {
				c.DataResponse = iris.Map{"code": -3, "msg": "Gem not enough"}
				return
			}
			wg.Add(1)
			go user.SetGem(gemFee*-1, "", 0, logtype.REMANE, ue.EventId, &wg)
		}
		//cập nhật số lần đổi
		ue.Turn++
		ue.UpdateDate = mysql.NullTime{Time: time.Now(), Valid: true}

		wg.Add(1)
		go func() {
			c.DB.Save(&ue)
			wg.Done()
		}()
		wg.Add(1)
		user.FullName = fullName
		go user.UpdateKey("FullName", user.FullName, &wg)
		wg.Wait()

		gemFee = ue.GetRenameFee(user)

		c.DataResponse = iris.Map{"code": 1, "msg": "Success", "data": iris.Map{"rename_fee": gemFee}}
	}
}

// /user/set/avatar
func (c *UserController) PostSetAvatar(form formValue, ctx iris.Context) {
	if c.validToken(ctx) {
		user := c.User

		avatarId := uint16(util.ToInt(form("avatar_id")))

		pet := models.HfPet{}
		_, check := pet.Find(avatarId, user)
		if !check {
			c.DataResponse = iris.Map{"code": -1, "msg": "Pet not Exists"}
			return
		}

		uPet := models.HfUserPet{}
		uPet, uCheck := uPet.Find(avatarId, user)
		if !uCheck || uPet.Stat == 0 {
			c.DataResponse = iris.Map{"code": -1, "msg": "Pet not own"}
			return
		}

		var wg sync.WaitGroup
		wg.Add(1)
		user.AvatarId = avatarId
		go user.UpdateKey("AvatarId", user.AvatarId, &wg)
		wg.Wait()

		c.DataResponse = iris.Map{"code": 1, "msg": "Success"}
	}
}
