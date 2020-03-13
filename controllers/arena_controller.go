package controllers

import (
	"GoLang/libraries/constants"
	"GoLang/libraries/logtype"
	"GoLang/libraries/util"
	"GoLang/models"
	"database/sql"
	"github.com/jinzhu/now"
	"github.com/kataras/iris"
	"math"
	"sort"
	"sync"
	"time"
)

type ArenaController struct {
	MyController
}

// /arena/info
func (c *ArenaController) GetInfo() {
	user := c.User

	uArena := models.HfUserArena{}
	uArena = uArena.Get(user)

	if uArena.RivalId != "" {
		uRival := models.HfUserArena{}
		uRival, findOk := uRival.Find(uArena.RivalId, c.DB)
		if findOk {
			uArena.EndgameWinLose(&uRival, false, "", user)
		}
	}

	myRank, tops := uArena.GetRankAndTop(c.DB)

	c.DataResponse = iris.Map{
		"code": 1,
		"data": iris.Map{
			"info":         uArena.GetMap(),
			"price_ticket": uArena.GetPriceTicket(1),
			"rank":         myRank,
			"top":          util.JsonDecodeArray(tops),
		},
	}
}

// /arena/award
func (c *ArenaController) GetAward() {

	aAward := models.HfArenaAward{}
	all := aAward.GetAll(c.User)

	sort.Slice(all, func(i, j int) bool {
		return all[i].Rank < all[j].Rank
	})

	boxDay := []iris.Map{}
	boxWeek := []iris.Map{}
	for _, val := range all {

		boxDay = append(boxDay, iris.Map{
			"rank":     val.Rank,
			"des_rank": val.DesRank,
			"gift":     util.JsonDecodeMap(val.BoxDay),
		})
		boxWeek = append(boxWeek, iris.Map{
			"rank":     val.Rank,
			"des_rank": val.DesRank,
			"gift":     util.JsonDecodeMap(val.BoxWeek),
		})
	}

	data := iris.Map{
		"box_day":  boxDay,
		"box_week": boxWeek,
	}

	c.DataResponse = iris.Map{
		"code": 1,
		"data": data,
	}
}

// /arena/find/rival
func (c *ArenaController) PostFindRival(ctx iris.Context) {
	if c.validToken(ctx) {
		user := c.User

		uArena := models.HfUserArena{}
		uArena = uArena.Get(user)

		if uArena.Ticket < 1 {
			c.DataResponse = iris.Map{"code": -1, "msg": "no tickets"}
			return
		}
		rivalId := uArena.RandomRival(user)

		rival := models.HfUserArena{}
		rival, find := rival.Find(rivalId, c.DB)

		if !find {
			c.DataResponse = iris.Map{"code": -2, "msg": "No rival found"}
			return
		}

		//trừ vé tham gia
		var wg sync.WaitGroup
		wg.Add(1)
		uArena.HitDaily++
		uArena.HitWeekly++
		go uArena.SetTicket(-1, "", nil, logtype.FIND_RIVAL_ARENA, 0, user, &wg)
		wg.Wait()

		c.DataResponse = iris.Map{"code": 1, "data": rival.GetMap()}
	}
}

// /arena/find/rival/v2
func (c *ArenaController) PostFindRivalV2(ctx iris.Context) {
	if c.validToken(ctx) {
		user := c.User

		uArena := models.HfUserArena{}
		uArena = uArena.Get(user)

		c.IsEncrypt = true
		if uArena.Ticket < 1 {
			c.DataResponse = iris.Map{"code": -1, "msg": "no tickets"}
			return
		}
		rivalId := uArena.RandomRival(user)

		rival := models.HfUserArena{}
		rival, find := rival.Find(rivalId, c.DB)

		if !find {
			c.DataResponse = iris.Map{"code": -2, "msg": "No rival found"}
			return
		}

		//trừ vé tham gia
		var wg sync.WaitGroup
		wg.Add(1)
		uArena.HitDaily++
		uArena.HitWeekly++
		go uArena.SetTicket(-1, "", nil, logtype.FIND_RIVAL_ARENA, 0, user, &wg)
		wg.Wait()

		c.DataResponse = iris.Map{"code": 1, "data": rival.GetMap()}
	}
}

// /arena/set/lineup
func (c *ArenaController) PostSetLineup(form formValue, ctx iris.Context) {
	if c.validToken(ctx) {
		user := c.User

		data := form("data")
		level := util.ParseInt(form("level"))
		power_point := util.ParseInt(form("power_point"))

		if !util.ValidHash(form, level, power_point, data) {
			c.DataResponse = iris.Map{"code": -1, "msg": "Invalid hash"}
			return
		}

		uArena := models.HfUserArena{}
		uArena = uArena.Get(user)

		uArena.LineUp = sql.NullString{String: data, Valid: true}
		uArena.Level = level
		uArena.PowerPoint = power_point
		c.DB.Save(&uArena)

		c.DataResponse = iris.Map{"code": 1, "msg": "Save success"}
	}
}

// /arena/buy/ticket
func (c *ArenaController) PostBuyTicket(form formValue, ctx iris.Context) {
	if c.validToken(ctx) {
		user := c.User

		quantity := math.Max(float64(util.ParseInt(form("quantity"))), 1)

		uArena := models.HfUserArena{}
		uArena = uArena.Get(user)

		priceTicket := uArena.GetPriceTicket(int(quantity))

		if user.Gem < uint(priceTicket) {
			c.DataResponse = iris.Map{"code": -1, "msg": "no gem"}
			return
		}

		gemFee := priceTicket * -1

		var wg sync.WaitGroup
		wg.Add(2)
		go uArena.SetTicket(int(quantity), "", 0, logtype.BUY_TICKET_ARENA, 0, user, &wg)
		go user.SetGem(gemFee, "", 0, logtype.BUY_TICKET_ARENA, 0, &wg)
		wg.Wait()

		c.DataResponse = iris.Map{"code": 1, "msg": "Buy success", "data": iris.Map{constants.GIFT: iris.Map{constants.TICKET_ARENA: quantity}, constants.GEM: user.Gem}}
	}
}

// /arena/endgame
func (c *ArenaController) PostEndgame(form formValue, ctx iris.Context) {
	if c.validToken(ctx) {
		user := c.User
		win := util.ParseInt(form("win"))
		user_id := form("user_id") //check có giống như đã tìm không
		data := form("data")

		isWin := win == 1 //1: thắng, 0: thua

		uArena := models.HfUserArena{}
		uArena = uArena.Get(user)
		uArena.Level = util.ParseInt(form("level"))
		uArena.PowerPoint = util.ParseInt(form("power_point"))
		uArena.LastLineUp = sql.NullString{String: form("last_line_up"), Valid: true}

		if uArena.RivalId == "" {
			c.DataResponse = iris.Map{"code": -1, "msg": "uArena.RivalId == ''"}
			return
		}

		if !util.ValidHash(form, win, uArena.Level, uArena.PowerPoint, uArena.LastLineUp.String) {
			c.DataResponse = iris.Map{"code": -1, "msg": "Invalid hash"}
			return
		}

		if user_id != uArena.RivalId {
			c.DataResponse = iris.Map{"code": -1, "msg": "user_id != uArena.RivalId"}
			return
		}

		uRival := models.HfUserArena{}
		uRival, findOk := uRival.Find(uArena.RivalId, c.DB)
		if !findOk {
			c.DataResponse = iris.Map{"code": -1, "msg": "can't find rival"}
			return
		}

		elo, gift := uArena.EndgameWinLose(&uRival, isWin, data, user)

		c.DataResponse = iris.Map{"code": 1, "msg": "Success", "data": iris.Map{"elo": uArena.Elo, "delta_elo": elo, constants.GIFT: gift}}
	}
}

// /arena/info/pvp
func (c *ArenaController) GetInfoPvp() {
	user := c.User

	uArena := models.HfUserArenaPvp{}
	uArena = uArena.Get(user)
	if uArena.LastWin != -1 {
		isWin := uArena.LastWin == 1 //1: thắng, 0: thua
		uArena.EndgameWinLose(isWin, "", user)
	}

	myRank, tops := uArena.GetRankAndTop(c.DB)
	endTime := now.New(now.Sunday()).EndOfDay().Unix() - time.Now().Unix()

	c.DataResponse = iris.Map{
		"code": 1,
		"data": iris.Map{
			"info":         uArena.GetMap(),
			"price_ticket": uArena.GetPriceTicket(),
			"rank":         myRank,
			"top":          util.JsonDecodeArray(tops),
			"time_reset":   endTime,
		},
	}
}

// /arena/pvp/award
func (c *ArenaController) GetPvpAward() {

	aAward := models.HfArenaPvpAward{}
	all := aAward.GetAll(c.User)

	sort.Slice(all, func(i, j int) bool {
		return all[i].Rank < all[j].Rank
	})

	boxWeek := []iris.Map{}
	for _, val := range all {

		boxWeek = append(boxWeek, iris.Map{
			"rank":     val.Rank,
			"des_rank": val.DesRank,
			"gift":     util.JsonDecodeMap(val.BoxWeek),
		})
	}

	data := iris.Map{
		"box_week": boxWeek,
	}

	c.DataResponse = iris.Map{
		"code": 1,
		"data": data,
	}
}

// /arena/buy/ticket/pvp
func (c *ArenaController) PostBuyTicketPvp(ctx iris.Context) {
	if c.validToken(ctx) {
		user := c.User

		uArena := models.HfUserArenaPvp{}
		uArena = uArena.Get(user)

		priceTicket := uArena.GetPriceTicket()

		if user.Gem < uint(priceTicket) {
			c.DataResponse = iris.Map{"code": -1, "msg": "no gem"}
			return
		}

		gemFee := priceTicket * -1

		var wg sync.WaitGroup
		wg.Add(2)
		go uArena.SetTicket(3, "", 0, logtype.BUY_TICKET_ARENA_PVP, 0, user, &wg)
		go user.SetGem(gemFee, "", 0, logtype.BUY_TICKET_ARENA_PVP, 0, &wg)
		wg.Wait()

		c.DataResponse = iris.Map{"code": 1, "msg": "Buy success", "data": iris.Map{constants.GIFT: iris.Map{constants.TICKET_ARENA_PVP: 3}, constants.GEM: user.Gem}}
	}
}

// /arena/endgame/pvp
func (c *ArenaController) PostEndgamePvp(form formValue, ctx iris.Context) {
	if c.validToken(ctx) {
		user := c.User
		data := form("data")

		uArena := models.HfUserArenaPvp{}
		uArena = uArena.Get(user)
		uArena.Level = util.ParseInt(form("level"))
		uArena.PowerPoint = util.ParseInt(form("power_point"))
		uArena.LastLineUp = sql.NullString{String: form("last_line_up"), Valid: true}
		uArena.HitWeekly++

		isWin := uArena.LastWin == 1 //1: thắng, 0: thua

		if uArena.LastWin == -1 {
			c.DataResponse = iris.Map{"code": -1, "msg": "Invalid uArena.LastWin == -1"}
			return
		}

		if !util.ValidHash(form, uArena.Level, uArena.PowerPoint, uArena.LastLineUp.String) {
			c.DataResponse = iris.Map{"code": -1, "msg": "Invalid hash"}
			return
		}

		elo, gift := uArena.EndgameWinLose(isWin, data, user)

		//Event
		ue := models.HfUserEvent{}
		ck, gf := ue.GiftPVP(user, isWin)
		if ck {
			var wg sync.WaitGroup
			wg.Add(1)
			go user.UpdateGifts(gf, logtype.GIFT_EVENT, ue.EventId, &wg)
			wg.Wait()

			gift = util.MergeGift(gift, gf)
		}

		c.DataResponse = iris.Map{"code": 1, "msg": "Success", "data": iris.Map{"elo_pvp": uArena.Elo, "delta_elo_pvp": elo, constants.GIFT: gift}}
	}
}

// /arena/info/shop
func (c *ArenaController) GetInfoShop() {
	user := c.User

	uShop := models.HfUserArenaShop{}
	uShop = uShop.Get(user)

	progressBought := map[int]int{}
	util.JsonDecodeObject(uShop.ItemsBought, &progressBought)

	aShop := models.HfArenaShop{}
	allShop := aShop.GetAll(user)

	itemPermanent := []iris.Map{}

	for _, shop := range allShop {
		bought := 0
		isBought := 0
		if val, ok := progressBought[shop.Id]; ok {
			bought = val
		}

		item := iris.Map{
			"id":         shop.Id,
			"item_id":    shop.Id,
			"name":       shop.Name,
			"price":      shop.Price,
			"cost":       shop.Cost,
			"price_type": shop.TypePrice,
			"limit":      shop.Limit,
			"bought":     bought,
			"is_bought":  isBought,
			"gift":       util.JsonDecodeMap(shop.Gift),
		}

		itemPermanent = append(itemPermanent, item)
	}

	arenaCoinResetConfig := uShop.CostResetConfig(user)
	timeReset := now.EndOfMonth().Unix() - time.Now().Unix()

	data := iris.Map{
		"permanent":               itemPermanent,
		"arena_coin_reset_config": arenaCoinResetConfig,
		"time_reset":              timeReset,
		constants.ARENA_COIN:      uShop.ArenaCoin,
	}

	c.DataResponse = iris.Map{"code": 1, "data": data}
}

// /arena/reset/shop
func (c *ArenaController) PostResetShop(form formValue, ctx iris.Context) {
	if c.validToken(ctx) {
		user := c.User

		uShop := models.HfUserArenaShop{}
		uShop = uShop.Get(user)

		arenaCoinResetConfig := uShop.CostResetConfig(user)

		if uShop.ArenaCoin < uint(arenaCoinResetConfig) {
			c.DataResponse = iris.Map{"code": -1, "msg": "Arena Coin not enough"}
			return
		}

		//Tiến hành reset
		uShop.ItemsBought = "{}"
		uShop.ResetDate = time.Now()
		//todo đã có lưu trong SetArenaCoin()

		//trừ coin
		var wg sync.WaitGroup
		wg.Add(1)
		costFee := arenaCoinResetConfig * -1
		go uShop.SetArenaCoin(costFee, "", 0, logtype.RESET_ARENA_SHOP, 0, user, &wg)
		wg.Wait()

		c.DataResponse = iris.Map{"code": 1, "msg": "Success"}
	}
}

// /arena/buy/shop
func (c *ArenaController) PostBuyShop(form formValue, ctx iris.Context) {
	if c.validToken(ctx) {
		user := c.User

		itemId := util.ToInt(form("item_id"))

		aShop := models.HfArenaShop{}
		aShop, check := aShop.Find(itemId, user)

		uShop := models.HfUserArenaShop{}
		uShop = uShop.Get(user)

		progressBought := map[int]int8{}
		util.JsonDecodeObject(uShop.ItemsBought, &progressBought)

		if !check {
			c.DataResponse = iris.Map{"code": -1, "msg": "item_id not found"}
			return
		}

		if val, ok := progressBought[itemId]; ok && val >= aShop.Limit {
			c.DataResponse = iris.Map{"code": -1, "msg": "Item Bought"}
			return
		}

		if uShop.ArenaCoin < aShop.Price {
			c.DataResponse = iris.Map{"code": -1, "msg": "Arena Coin not enough"}
			return
		}

		var wg sync.WaitGroup
		wg.Add(1)
		costFee := int(aShop.Price) * -1
		go uShop.SetArenaCoin(costFee, "", itemId, logtype.BUY_ARENA_SHOP, 0, user, &wg)
		wg.Wait()

		//đánh dấu mua, cập nhật quà
		gifts := util.JsonDecodeMap(aShop.Gift)

		wg.Add(1)
		go user.UpdateGifts(gifts, logtype.BUY_ARENA_SHOP, 0, &wg)

		if _, ok := progressBought[itemId]; ok {
			progressBought[itemId]++
		} else {
			progressBought[itemId] = 1
		}

		uShop.ItemsBought = util.JsonEndCode(progressBought)
		wg.Add(1)
		go func() {
			user.DB.Save(&uShop)
			wg.Done()
		}()

		wg.Add(1)
		go uShop.SaveLog(user, itemId, gifts, &wg)

		wg.Wait()
		c.DataResponse = iris.Map{"code": 1, "data": iris.Map{constants.GIFT: gifts}}
	}
}
