package controllers

import (
	"GoLang/config/cashshop"
	"GoLang/libraries/constants"
	"GoLang/libraries/logtype"
	"GoLang/libraries/util"
	"GoLang/models"
	"database/sql"
	"github.com/go-sql-driver/mysql"
	"github.com/jinzhu/now"
	"github.com/kataras/iris"
	"github.com/mileusna/useragent"
	"math"
	"strings"
	"sync"
	"time"
)

type CashshopController struct {
	MyController
}

// /cashshop/info
func (c *CashshopController) GetInfo() {
	user := c.User
	uCashshop := models.HfUserCashshop{}
	uCashshop.Get(user)
	firstPackage := map[string]uint8{}
	util.JsonDecodeObject(uCashshop.FirstPackage, &firstPackage)

	data := []iris.Map{}
	csConfig := cashshop.Config()

	for id, conf := range csConfig {
		if id == "p7" {
			continue
		}

		if c.UserAgent.OS == ua.Android && (id == "p2" || id == "p3" || id == "p4" || id == "p5" || id == "p6") {
			continue
		}

		dataMap := iris.Map{
			"package_id":    id,
			"package_name":  conf.PackageName,
			"web":           conf.Web,
			"google":        conf.Google,
			"ios":           conf.Ios,
			"gem":           conf.Gem,
			"gem_first":     conf.GemFirst,
			"first_package": util.ToInt(firstPackage[id]),
		}

		if id == "p8" {
			ue := models.HfUserEvent{}
			ue, check := ue.Get(8, user)
			dataMap["active"] = false
			dataMap["is_receive"] = now.New(ue.UpdateDate.Time).EndOfDay() == now.EndOfDay()
			dataMap["end_day"] = math.Max(0, float64(ue.ReceiveDate.Time.Unix()-time.Now().Unix()))
			if check && ue.ReceiveDate.Time.Unix() > time.Now().Unix() {
				dataMap["active"] = true
			}
		}

		if id == "p9" {
			ue := models.HfUserEvent{}
			ue, check := ue.Get(9, user)
			dataMap["active"] = false
			dataMap["is_receive"] = now.New(ue.UpdateDate.Time).EndOfDay() == now.EndOfDay()
			dataMap["end_day"] = math.Max(0, float64(ue.ReceiveDate.Time.Unix()-time.Now().Unix()))
			if check && ue.ReceiveDate.Time.Unix() > time.Now().Unix() {
				dataMap["active"] = true
			}
		}

		data = append(data, dataMap)
	}

	c.DataResponse = iris.Map{"code": 1, "data": data}
}

// /cashshop/generate/orderid
func (c *CashshopController) PostGenerateOrderid(form formValue, ctx iris.Context) {
	if c.validToken(ctx) {
		user := c.User

		PackageID := form("package_id")

		csConfig := cashshop.Config()
		if _, ok := csConfig[PackageID]; !ok {
			c.DataResponse = iris.Map{"code": -1, "msg": "PackageID not Exists"}
			return
		}

		uTran := models.HfUserTransaction{}
		uTran.OrderId = util.UUID()
		uTran.UserId = user.UserId
		uTran.ServerId = user.ServerId
		uTran.PackageId = PackageID
		uTran.TimeOrder = time.Now()
		uTran.Type = constants.PAYMENT_INGAME

		c.DB.Save(&uTran)

		c.DataResponse = iris.Map{"code": 1, "data": iris.Map{"game_order_id": uTran.OrderId}}
	}
}

// /cashshop/gift/week
func (c *CashshopController) PostGiftWeek(form formValue, ctx iris.Context) {
	eventId := 8
	packageId := "p8"
	if c.validToken(ctx) {
		user := c.User
		ue := models.HfUserEvent{}
		ue, check := ue.Get(eventId, user)

		//Không có data hoặc đã nhận hôm nay hoặc đã hết hạn gói
		if !check || now.New(ue.UpdateDate.Time).EndOfDay() == now.EndOfDay() || ue.ReceiveDate.Time.Unix() < time.Now().Unix() {
			c.DataResponse = iris.Map{"code": -1, "msg": "Invalid"}
			return
		}

		conf := models.HfConfig{}
		giftGem, _ := conf.GetPackageGift(packageId, user)

		//T7 CN x2
		if time.Now().Weekday() == time.Saturday || time.Now().Weekday() == time.Sunday {
			giftGem *= 2
		}
		gifts := map[string]interface{}{constants.GEM: giftGem}

		//Tiến hành nhận quà
		var wg sync.WaitGroup
		wg.Add(1)
		go user.UpdateGifts(gifts, logtype.GIFT_EVENT, eventId, &wg)

		//Cập nhật trạng thái + log
		ue.UpdateDate = mysql.NullTime{Time: time.Now(), Valid: true}
		wg.Add(1)
		go func() {
			c.DB.Save(&ue)
			wg.Done()
		}()

		wg.Add(1)
		go ue.SaveLog(user, 1, 0, gifts, &wg)
		wg.Wait()

		c.DataResponse = iris.Map{"code": 1, "data": iris.Map{constants.GIFT: gifts}}
	}
}

// /cashshop/gift/month
func (c *CashshopController) PostGiftMonth(form formValue, ctx iris.Context) {
	eventId := 9
	packageId := "p9"
	if c.validToken(ctx) {
		user := c.User
		ue := models.HfUserEvent{}
		ue, check := ue.Get(eventId, user)

		//Không có data hoặc đã nhận hôm nay hoặc đã hết hạn gói
		if !check || now.New(ue.UpdateDate.Time).EndOfDay() == now.EndOfDay() || ue.ReceiveDate.Time.Unix() < time.Now().Unix() {
			c.DataResponse = iris.Map{"code": -1, "msg": "Invalid"}
			return
		}

		conf := models.HfConfig{}
		giftGem, _ := conf.GetPackageGift(packageId, user)

		//T7 CN x2
		if time.Now().Weekday() == time.Saturday || time.Now().Weekday() == time.Sunday {
			giftGem *= 2
		}
		gifts := map[string]interface{}{constants.GEM: giftGem}

		//Tiến hành nhận quà
		var wg sync.WaitGroup
		wg.Add(1)
		go user.UpdateGifts(gifts, logtype.GIFT_EVENT, eventId, &wg)

		//Cập nhật trạng thái + log
		ue.UpdateDate = mysql.NullTime{Time: time.Now(), Valid: true}
		wg.Add(1)
		go func() {
			c.DB.Save(&ue)
			wg.Done()
		}()

		wg.Add(1)
		go ue.SaveLog(user, 1, 0, gifts, &wg)
		wg.Wait()

		c.DataResponse = iris.Map{"code": 1, "data": iris.Map{constants.GIFT: gifts}}
	}
}

// /cashshop/info/exchange/gold
func (c *CashshopController) GetInfoExchangeGold() {
	user := c.User

	ue := models.HfUserEvent{}
	ue, check := ue.Get(6, user)
	info := []iris.Map{{"id": 1, "gem": 30, "gold": 20000, "limit": 10, "exchange": 0}}

	// nếu có trong ngày check số lần đổi
	if check && now.New(ue.UpdateDate.Time).BeginningOfDay() == now.BeginningOfDay() {

		progress := map[int]int{}

		util.JsonDecodeObject(ue.Progress.String, &progress)

		info = []iris.Map{{"id": 1, "gem": 30, "gold": 20000, "limit": 10, "exchange": progress[1]}}
	}
	c.DataResponse = iris.Map{"code": 1, "data": info}
}

// /cashshop/exchange/gold
func (c *CashshopController) PostExchangeGold(form formValue, ctx iris.Context) {
	if c.validToken(ctx) {
		user := c.User

		id := util.ToInt(form("id"))
		if !util.InArray(id, []int{1}) {
			c.DataResponse = iris.Map{"code": -1, "msg": "Id invalid"}
			return
		}

		price := map[int]uint{1: 30}
		gold := map[int]int{1: 20000}
		if user.Gem < price[id] {
			c.DataResponse = iris.Map{"code": -1, "msg": "Gem not enough"}
			return
		}

		ue := models.HfUserEvent{}
		ue, check := ue.Get(6, user)

		// nếu có trong ngày check số lần đổi
		progress := map[int]int{1: 0}
		if check {

			if now.New(ue.UpdateDate.Time).BeginningOfDay() == now.BeginningOfDay() {
				util.JsonDecodeObject(ue.Progress.String, &progress)

				if progress[id] >= 10 {
					c.DataResponse = iris.Map{"code": -1, "msg": "Het luot doi"}
					return
				}
			} else {
				ue.UpdateDate = mysql.NullTime{Time: time.Now(), Valid: true}
			}
		} else {
			ue.UpdateDate = mysql.NullTime{Time: time.Now(), Valid: true}
		}
		var wg sync.WaitGroup

		progress[id]++
		ue.Progress = sql.NullString{String: util.JsonEndCode(progress), Valid: true}

		wg.Add(1)
		go func() {
			c.DB.Save(&ue)
			wg.Done()
		}()

		gift := iris.Map{constants.GOLD: gold[id]}

		wg.Add(1)
		gemFee := int(price[id]) * -1
		go user.SetGem(gemFee, "", uint(id), logtype.EXCHANGE_GOLD, ue.EventId, &wg)

		wg.Add(1)
		go user.SetGold(gold[id], "", uint(id), logtype.EXCHANGE_GOLD, ue.EventId, &wg)

		wg.Add(1)
		go ue.SaveLog(user, 0, uint(id), gift, &wg)

		wg.Add(1)
		go user.CompleteMissionDaily(models.QDL_BUY_GOLD, &wg)

		wg.Wait()

		c.DataResponse = iris.Map{"code": 1, "data": iris.Map{constants.GIFT: gift}}
	}
}

// /cashshop/info/market/black
func (c *CashshopController) GetInfoMarketBlack() {
	user := c.User

	con := models.HfConfig{}
	timeResetConfig := util.ToInt(con.Find("market_black_time_reset", user).Value)

	con = models.HfConfig{}
	marketBlackMaxTicketReset := util.ToInt(con.Find("market_black_max_ticket_reset", user).Value)

	con = models.HfConfig{}
	gemResetConfig := util.ToInt(con.Find("market_black_gem_reset", user).Value)

	ticketMarketBlack := 0
	uItem, check := user.GetItem(constants.TICKET_MARKET_BLACK)
	var wg sync.WaitGroup
	if check {
		ticketMarketBlack = uItem.Quantity
	} else {
		ticketMarketBlack = marketBlackMaxTicketReset
		wg.Add(1)
		uItem.TimeCheck = mysql.NullTime{Time: time.Now(), Valid: true}
		go uItem.Set(marketBlackMaxTicketReset, logtype.GET_INFO_TICKET_MARKET_BLACK, 0, user, &wg)
		wg.Wait()
	}

	//cộng lượt
	ticketPlus := util.ToInt(util.ToInt(time.Now().Unix()-uItem.TimeCheck.Time.Unix()) / timeResetConfig)
	if ticketPlus > 0 {
		if ticketMarketBlack+ticketPlus > marketBlackMaxTicketReset {
			ticketPlus = marketBlackMaxTicketReset - ticketMarketBlack
		}
		ticketMarketBlack += ticketPlus
		uItem.TimeCheck = mysql.NullTime{Time: time.Now(), Valid: true}

		if ticketPlus > 0 {
			wg.Add(1)
			go uItem.Set(ticketPlus, logtype.GET_INFO_TICKET_MARKET_BLACK, 0, user, &wg)
			wg.Wait()
		}
	}

	timeReset := uItem.TimeCheck.Time.Unix() + int64(timeResetConfig) - time.Now().Unix()

	ue := models.HfUserEvent{}
	ue = ue.GetMarketBlack(user)

	progressBonus := util.JsonDecodeMap(ue.ProgressBonus.String)

	idItemRans := map[int]int{}
	progressBought := map[int]int{}
	resetBought := map[int]int{}

	util.JsonDecodeObject(ue.Progress.String, &idItemRans)
	util.JsonDecodeObject(util.JsonEndCode(progressBonus["limit"]), &progressBought)
	util.JsonDecodeObject(util.JsonEndCode(progressBonus["reset"]), &resetBought)

	shop := models.HfShop{}
	allShop := shop.GetAll(user)

	itemRandom := []iris.Map{}
	itemPermanent := []iris.Map{}

	for id, itemId := range idItemRans {
		for _, shop := range allShop {
			if itemId == shop.Id {
				bought := 0
				isBought := 0
				if val, ok := progressBought[shop.Id]; ok {
					bought = val
				}
				if val, ok := resetBought[id]; ok {
					isBought = val
				}
				//nếu đã mua trong lần reset và trong ngày chưa mua
				if bought < isBought {
					bought = isBought
				}

				item := iris.Map{
					"id":         shop.Id,
					"item_id":    id,
					"name":       shop.Name,
					"price":      shop.Price,
					"cost":       shop.Cost,
					"price_type": shop.TypePrice,
					"limit":      shop.Limit,
					"bought":     bought,
					"is_bought":  isBought,
					"gift":       util.JsonDecodeMap(shop.Gift),
				}

				if util.InArray(shop.Id, []int{34, 35, 36}) {
					levelEquip := int(ue.Turn) + (shop.Id - 34)
					item["price"] = uint(ue.Turn) * shop.Price

					gift := strings.Replace(shop.Gift, `"?"`, util.ToString(levelEquip), -1)
					item["gift"] = util.JsonDecodeMap(gift)
				}
				itemRandom = append(itemRandom, item)

				break
			}
		}
	}

	for _, shop := range allShop {
		if shop.IsRandom == 0 {
			bought := 0
			isBought := 0
			if val, ok := progressBought[shop.Id]; ok {
				bought = val
			}
			//if val, ok := resetBought[shop.Id]; ok {
			//	isBought = val
			//}

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

	}

	data := iris.Map{
		"permanent":                   itemPermanent,
		"random":                      itemRandom,
		"time_reset_config":           timeResetConfig,
		"gem_reset_config":            gemResetConfig,
		"time_reset":                  timeReset,
		constants.TICKET_MARKET_BLACK: ticketMarketBlack,
	}

	c.DataResponse = iris.Map{"code": 1, "data": data}

	//shop.RandomMarketBlack(user, ue.ProgressBonus.String)
}

// /cashshop/reset/market/black
func (c *CashshopController) PostResetMarketBlack(form formValue, ctx iris.Context) {
	if c.validToken(ctx) {
		user := c.User
		uItem, _ := user.GetItem(constants.TICKET_MARKET_BLACK)

		con := models.HfConfig{}
		gemResetConfig := util.ToInt(con.Find("market_black_gem_reset", user).Value)

		if uItem.Quantity <= 0 && user.Gem < uint(gemResetConfig) {
			c.DataResponse = iris.Map{"code": -1, "msg": "Don't have gem or ticket"}
			return
		}

		//check để trừ vé or gem
		var wg sync.WaitGroup
		if uItem.Quantity != 0 {
			con = models.HfConfig{}
			marketBlackMaxTicketReset := util.ToInt(con.Find("market_black_max_ticket_reset", user).Value)
			if uItem.Quantity >= marketBlackMaxTicketReset {
				uItem.TimeCheck = mysql.NullTime{Time: time.Now(), Valid: true}
			}

			wg.Add(1)
			go uItem.Set(-1, logtype.RESET_MARKET_TICKET_MARKET_BLACK, 0, user, &wg)
		} else {
			wg.Add(1)
			gemFee := gemResetConfig * -1
			go user.SetGem(gemFee, "", 0, logtype.RESET_MARKET_TICKET_MARKET_BLACK, 0, &wg)
		}
		wg.Wait()

		//Tiến hành reset
		ue := models.HfUserEvent{}
		ue = ue.GetMarketBlack(user)
		shop := models.HfShop{}
		idItems, levelEquip, progressBonus := shop.RandomMarketBlack(user, ue.ProgressBonus.String)

		ue.Progress = sql.NullString{String: util.JsonEndCode(idItems), Valid: true}
		ue.Turn = uint(levelEquip)

		progressBonus["reset"] = iris.Map{}

		ue.ProgressBonus = sql.NullString{String: util.JsonEndCode(progressBonus), Valid: true}

		ue.UpdateDate = mysql.NullTime{Time: time.Now(), Valid: true}
		user.DB.Save(&ue)

		c.DataResponse = iris.Map{"code": 1, "msg": "Success"}
	}
}

// /cashshop/buy/market/black
func (c *CashshopController) PostBuyMarketBlack(form formValue, ctx iris.Context) {
	if c.validToken(ctx) {
		user := c.User

		itemId := util.ToInt(form("item_id"))

		if itemId > 10 && itemId < 101 {
			c.DataResponse = iris.Map{"code": -1, "msg": "item_id invalid"}
			return
		}

		ue := models.HfUserEvent{}
		ue = ue.GetMarketBlack(user)

		progressBonus := util.JsonDecodeMap(ue.ProgressBonus.String)

		idItemRans := map[int]int{}
		progressBought := map[int]int{}
		resetBought := map[int]int{}
		freeBought := map[int]int{}

		util.JsonDecodeObject(ue.Progress.String, &idItemRans)
		util.JsonDecodeObject(util.JsonEndCode(progressBonus["limit"]), &progressBought)
		util.JsonDecodeObject(util.JsonEndCode(progressBonus["reset"]), &resetBought)
		util.JsonDecodeObject(util.JsonEndCode(progressBonus["free"]), &freeBought)

		if _, ok := resetBought[itemId]; ok && itemId <= 10 {
			c.DataResponse = iris.Map{"code": -1, "msg": "Item Bought"}
			return
		}

		id := itemId
		if id <= 10 {
			id = idItemRans[itemId]
		}
		shop := models.HfShop{}
		shop, check := shop.Find(id, user)
		if !check {
			c.DataResponse = iris.Map{"code": -1, "msg": "Can't found item"}
			return
		}

		if bought, ok := progressBought[id]; ok && shop.Limit != -1 {
			if bought >= int(shop.Limit) {
				c.DataResponse = iris.Map{"code": -1, "msg": "Item limit"}
				return
			}
		}

		//kiểm tra tiền, quà
		if util.InArray(shop.Id, []int{34, 35, 36}) {
			levelEquip := uint(ue.Turn) + uint(shop.Id-34)
			shop.Price *= levelEquip

			shop.Gift = strings.Replace(shop.Gift, `"?"`, util.ToString(levelEquip), -1)
		}

		var wg sync.WaitGroup

		if shop.Price > 0 {
			switch shop.TypePrice {
			case constants.GOLD:
				if user.Gold < shop.Price {
					c.DataResponse = iris.Map{"code": -1, "msg": "Gold not enough"}
					return
				}
				wg.Add(1)
				goldFee := int(shop.Price) * -1
				go user.SetGold(goldFee, "", uint(shop.Id), logtype.BUY_MARKET_BLACK, ue.EventId, &wg)

				wg.Add(1)
				go user.CompleteMissionDaily(models.QDL_MARKET_GOLD, &wg)

			case constants.GEM:
				if user.Gem < shop.Price {
					c.DataResponse = iris.Map{"code": -1, "msg": "Gem not enough"}
					return
				}
				wg.Add(1)
				gemFee := int(shop.Price) * -1
				go user.SetGem(gemFee, "", uint(shop.Id), logtype.BUY_MARKET_BLACK, ue.EventId, &wg)

				wg.Add(1)
				go user.CompleteMissionDaily(models.QDL_MARKET_GEM, &wg)

			}
		}
		wg.Wait()

		//đánh dấu mua, cập nhật quà
		gifts := map[string]interface{}{}
		gifts_cf := util.JsonDecodeMap(shop.Gift)
		if val, ok := gifts_cf[constants.RANDOM]; ok {
			gifts = util.RandomPercentGift(val)
		} else {
			gifts = gifts_cf
		}

		wg.Add(1)
		go user.UpdateGifts(gifts, logtype.BUY_MARKET_BLACK, ue.EventId, &wg)

		if _, ok := progressBought[id]; ok {
			progressBought[id]++
		} else {
			progressBought[id] = 1
		}
		if _, ok := resetBought[itemId]; ok {
			resetBought[itemId]++
		} else {
			resetBought[itemId] = 1
		}

		if shop.Price == 0 {
			freeBought[id]++
		}

		progressBonus["limit"] = progressBought
		progressBonus["reset"] = resetBought
		progressBonus["free"] = freeBought

		ue.ProgressBonus = sql.NullString{String: util.JsonEndCode(progressBonus), Valid: true}
		wg.Add(1)
		go func() {
			user.DB.Save(&ue)
			wg.Done()
		}()
		wg.Add(1)
		go ue.SaveLog(user, uint(itemId), uint(id), gifts, &wg)

		wg.Wait()
		c.DataResponse = iris.Map{"code": 1, "data": iris.Map{constants.GIFT: gifts}}
	}
}
