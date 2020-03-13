package controllers

import (
	"GoLang/libraries/constants"
	"GoLang/libraries/logtype"
	"GoLang/libraries/util"
	"GoLang/models"
	"database/sql"
	"github.com/go-sql-driver/mysql"
	"github.com/jinzhu/now"
	"github.com/kataras/iris"
	"sync"
	"time"
)

type EventController struct {
	MyController
}

// /event/all
func (c *EventController) GetAll() {
	user := c.User

	ev := models.HfEvent{}
	evs := ev.GetAll(user)
	data := []iris.Map{}

	ue := models.HfUserEvent{}

	sv := models.HfServer{}
	sv, check := sv.Find(user.ServerId, user)
	if !check {
		sv.Find(0, user)
	}
	timeOpen := sv.DateOpen

	//dateOpen := user.GetConfig("date_open_server_" + util.ToString(user.ServerId)).Value
	//timeOpen := now.MustParse(dateOpen)

	for _, val := range evs {

		//Kiểm tra có bị ẩn không
		if val.Id != 4 && (util.InArray(c.UserAgent.OS, util.JsonDecodeArray(val.DeviceHide)) || c.UserVersion > c.ServerVersion) {
			continue
		}

		//Todo không có ngày là vĩnh viễn, có ngày thì set 4,13,14,15 hoặc event còn hạn
		if !val.StartDate.Valid || val.StartDate.Time.Unix() <= time.Now().Unix() && val.EndDate.Time.Unix() >= time.Now().Unix() {
			switch val.Id {
			case 4:
				//continue
				val.StartDate.Time = user.CreatedDate
				val.EndDate.Time = now.New(now.New(user.CreatedDate).AddDate(0, 0, 6)).EndOfDay()
				dataNewbie := val.GetMap(user)
				if val.EndDate.Time.Unix() > time.Now().Unix() || dataNewbie["is_notify"] == 1 {
					data = append(data, dataNewbie)
				}
			case 7:
				ue := models.HfUserEvent{}
				ue, checkFirst := ue.Get(7, user)
				if !checkFirst {
					data = append(data, val.GetMap(user))
				} else {
					progress := map[uint16]int{}
					util.JsonDecodeObject(ue.Progress.String, &progress)
					for _, v := range progress {
						if v != 1 {
							data = append(data, val.GetMap(user))
							break
						}
					}
				}
			case 13:
				endTime := ue.EndTimeOpenDays(user, 7)
				if endTime > 0 {
					val.StartDate.Time = timeOpen
					val.EndDate.Time = time.Now().Add(time.Second * time.Duration(endTime))
					if val.EndDate.Time.Unix() > time.Now().Unix() {
						data = append(data, val.GetMap(user))
					}
				}
			case 14:
				endTime := ue.EndTimeOpenDays(user, 7)
				if endTime > 0 {
					val.StartDate.Time = timeOpen
					val.EndDate.Time = time.Now().Add(time.Second * time.Duration(endTime))
					if val.EndDate.Time.Unix() > time.Now().Unix() {
						data = append(data, val.GetMap(user))
					}
				}
			case 15:
				endTime := ue.EndTimeOpenDays(user, 4)
				if endTime > 0 {
					val.StartDate.Time = timeOpen
					val.EndDate.Time = time.Now().Add(time.Second * time.Duration(endTime))
					if val.EndDate.Time.Unix() > time.Now().Unix() {
						data = append(data, val.GetMap(user))
					}
				}

			default:
				data = append(data, val.GetMap(user))
			}
		}
	}

	c.DataResponse = iris.Map{"code": 1, "data": data, "client":c.UserVersion, "server":c.ServerVersion, "os":c.UserAgent.OS}
}

// /event/info/firstpayment
func (c *EventController) GetInfoFirstpayment() {
	user := c.User

	ue := models.HfUserEvent{}
	ue, check := ue.Get(7, user)

	event := models.HfEvent{}
	event, _ = event.Find(7, user)

	payment := uint(0)
	progress := map[uint16]int{0: 2, 500: 2, 1000: 2}

	if check {
		util.JsonDecodeObject(ue.Progress.String, &progress)
		payment = ue.Turn
	}

	c.DataResponse = iris.Map{"code": 1, "data": iris.Map{"payment": payment, "progress": progress}}
	if util.InArray(c.UserAgent.OS,util.JsonDecodeArray(event.DeviceHide)) || c.UserVersion > c.ServerVersion{
		c.DataResponse = iris.Map{"code": -1, "data": iris.Map{"payment": payment, "progress": progress}}
	}
}

// /event/gift/firstpayment
func (c *EventController) PostGiftFirstpayment(form formValue, ctx iris.Context) {
	if c.validToken(ctx) {
		user := c.User

		id := uint(util.ToInt(form("id")))

		ue := models.HfUserEvent{}
		ue, check := ue.Get(7, user)

		payment := uint(0)
		progress := map[uint]int{}

		if !check {
			c.DataResponse = iris.Map{"code": -1, "msg": "Invalid"}
			return
		}
		util.JsonDecodeObject(ue.Progress.String, &progress)
		payment = ue.Turn

		if val, ok := progress[id]; !ok || val != 0 {
			c.DataResponse = iris.Map{"code": -1, "msg": "Invalid Gift"}
			return
		}

		gift_cf := map[uint]iris.Map{
			0: {
				constants.PIECE:  map[int]int{3030: 30},
				constants.GEM:    888,
				constants.GOLD:   88888,
				constants.STONES: iris.Map{"evo": 188},
			},
			500: {
				constants.PIECE: map[int]int{10: 30},
				constants.EQUIP: map[int]int{110: 1, 210: 1},
			},
			1000: {
				constants.PIECE: map[int]int{3004: 50},
				constants.EQUIP: map[int]int{310: 1, 410: 1, 510: 1, 610: 1},
			},
		}

		//đánh dấu nhận
		progress[id] = 1
		gifts := util.InterfaceToMap(gift_cf[id])

		ue.Progress = sql.NullString{String: util.JsonEndCode(progress), Valid: true}

		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			c.DB.Save(&ue)
			wg.Done()
		}()

		wg.Add(1)
		go user.UpdateGifts(gifts, logtype.GIFT_EVENT, ue.EventId, &wg)

		wg.Add(1)
		go ue.SaveLog(user, 0, id, gifts, &wg)

		wg.Wait()

		c.DataResponse = iris.Map{"code": 1, "data": iris.Map{"payment": payment, "progress": progress, constants.GIFT: gifts}}
	}
}

// /event/info/grow/up
func (c *EventController) GetInfoGrowUp() {
	user := c.User
	ue := models.HfUserEvent{}
	ue = ue.GetGrowUp(user)

	progress := map[int]interface{}{}
	progressBonus := map[int]interface{}{}
	util.JsonDecodeObject(ue.Progress.String, &progress)
	util.JsonDecodeObject(ue.ProgressBonus.String, &progressBonus)
	isActive := ue.Type

	gift_cf := map[int]int{5: 50, 10: 50, 15: 50, 20: 50, 25: 50, 30: 50, 35: 50, 40: 50, 45: 50, 50: 100}
	gift_bonus_cf := map[int]int{5: 500, 10: 550, 15: 600, 20: 650, 25: 700, 30: 800, 35: 900, 40: 1000, 45: 1100, 50: 1200}

	for lv, val := range progress {
		progress[lv] = iris.Map{"status": util.StatusGift(val), constants.GIFT: iris.Map{constants.GEM: gift_cf[lv]}}
	}

	for lv, val := range progressBonus {
		progressBonus[lv] = iris.Map{"status": util.StatusGift(val), constants.GIFT: iris.Map{constants.GEM: gift_bonus_cf[lv]}}
	}

	c.DataResponse = iris.Map{"code": 1, "data": iris.Map{
		"progress":       progress,
		"progress_bonus": progressBonus,
		"is_active":      isActive,
	}}
}

// /event/gift/grow/up
func (c *EventController) PostGiftGrowUp(form formValue, ctx iris.Context) {
	if c.validToken(ctx) {
		user := c.User
		ue := models.HfUserEvent{}
		ue = ue.GetGrowUp(user)

		level := util.ToInt(form("level"))
		isBonus := uint(util.ToInt(form("is_bonus")))

		progress := map[int]int{}
		gift_cf := map[int]int{}
		if isBonus == 1 {
			util.JsonDecodeObject(ue.ProgressBonus.String, &progress)
			gift_cf = map[int]int{5: 500, 10: 550, 15: 600, 20: 650, 25: 700, 30: 800, 35: 900, 40: 1000, 45: 1100, 50: 1200}
		} else {
			util.JsonDecodeObject(ue.Progress.String, &progress)
			gift_cf = map[int]int{5: 50, 10: 50, 15: 50, 20: 50, 25: 50, 30: 50, 35: 50, 40: 50, 45: 50, 50: 100}
		}

		isActive := ue.Type
		if isBonus == 1 && isActive != 1 {
			c.DataResponse = iris.Map{"code": -1, "msg": "Grow up package not active"}
			return
		}

		if val, ok := progress[level]; !ok || val != 0 {
			c.DataResponse = iris.Map{"code": -1, "msg": "Gift invalid"}
			return
		}

		if _, ok := gift_cf[level]; !ok {
			c.DataResponse = iris.Map{"code": -1, "msg": "Can't found gift"}
			return
		}

		gifts := map[string]interface{}{constants.GEM: gift_cf[level]}

		//đánh dấu, nhận quà, lưu log
		progress[level] = 1
		if isBonus == 1 {
			ue.ProgressBonus = sql.NullString{String: util.JsonEndCode(progress), Valid: true}
		} else {
			ue.Progress = sql.NullString{String: util.JsonEndCode(progress), Valid: true}
		}
		ue.UpdateDate = mysql.NullTime{Time: time.Now(), Valid: true}

		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			c.DB.Save(&ue)
			wg.Done()
		}()

		wg.Add(1)
		go user.UpdateGifts(gifts, logtype.GIFT_EVENT, ue.EventId, &wg)

		wg.Add(1)
		go ue.SaveLog(user, isBonus, uint(level), gifts, &wg)
		wg.Wait()

		c.DataResponse = iris.Map{"code": 1, "data": iris.Map{constants.GIFT: gifts, "level": level, "is_bonus": isBonus}}
	}
}

// /event/info/newbie/rollup
func (c *EventController) GetInfoNewbieRollup() {
	user := c.User
	ue := models.HfUserEvent{}
	ue = ue.GetNewbieRollup(user)

	event := models.HfEvent{}
	event, _ = event.Find(ue.EventId, user)

	progress := map[int]interface{}{}
	util.JsonDecodeObject(ue.Progress.String, &progress)

	gift_cf := ue.GetNewbieRollupGift()

	for d, val := range progress {
		progress[d] = iris.Map{"status": util.StatusGift(val), constants.GIFT: gift_cf[d]}
	}

	c.DataResponse = iris.Map{"code": 1, "data": iris.Map{"progress": progress}}
	if util.InArray(c.UserAgent.OS,util.JsonDecodeArray(event.DeviceHide)) || c.UserVersion > c.ServerVersion{
		c.DataResponse = iris.Map{"code": -1, "data": iris.Map{"progress": progress}}
	}
}

// /event/gift/newbie/rollup
func (c *EventController) PostGiftNewbieRollup(form formValue, ctx iris.Context) {
	if c.validToken(ctx) {
		user := c.User
		ue := models.HfUserEvent{}
		ue = ue.GetNewbieRollup(user)

		day := util.ToInt(form("day"))

		progress := map[int]int{}
		util.JsonDecodeObject(ue.Progress.String, &progress)

		if val, ok := progress[day]; !ok || val != 0 {
			c.DataResponse = iris.Map{"code": -1, "msg": "Gift invalid"}
			return
		}

		gift_cf := ue.GetNewbieRollupGift()

		gifts := util.InterfaceToMap(gift_cf[day])
		//đánh dấu, nhận quà, lưu log
		progress[day] = 1
		ue.Progress = sql.NullString{String: util.JsonEndCode(progress), Valid: true}
		ue.UpdateDate = mysql.NullTime{Time: time.Now(), Valid: true}

		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			c.DB.Save(&ue)
			wg.Done()
		}()

		wg.Add(1)
		go user.UpdateGifts(gifts, logtype.GIFT_EVENT, ue.EventId, &wg)

		wg.Add(1)
		go ue.SaveLog(user, 0, uint(day), gifts, &wg)
		wg.Wait()

		c.DataResponse = iris.Map{"code": 1, "data": iris.Map{constants.GIFT: gifts}}
	}
}

// /event/info/payment/everyday
func (c *EventController) GetInfoPaymentEveryday() {
	user := c.User
	ue := models.HfUserEvent{}
	ue = ue.GetPaymentEveryday(user)

	progress := map[uint]interface{}{}
	util.JsonDecodeObject(ue.Progress.String, &progress)

	gift_cf := ue.GetPaymentEverydayGift()

	for g, val := range progress {
		progress[g] = iris.Map{"status": util.StatusGift(val), constants.GIFT: gift_cf[g]}
	}

	c.DataResponse = iris.Map{"code": 1, "data": iris.Map{
		"progress":   progress,
		"payment":    ue.Turn,
		"time_reset": now.EndOfDay().Unix() - time.Now().Unix(),
	}}
}

// /event/gift/payment/everyday
func (c *EventController) PostGiftPaymentEveryday(form formValue, ctx iris.Context) {
	if c.validToken(ctx) {
		user := c.User
		ue := models.HfUserEvent{}
		ue = ue.GetPaymentEveryday(user)

		typeGift := uint(util.ToInt(form("type")))

		progress := map[uint]int{}
		util.JsonDecodeObject(ue.Progress.String, &progress)

		if val, ok := progress[typeGift]; !ok || val != 0 {
			c.DataResponse = iris.Map{"code": -1, "msg": "Gift invalid"}
			return
		}

		gift_cf := ue.GetPaymentEverydayGift()

		gifts := util.InterfaceToMap(gift_cf[typeGift])
		progress[typeGift] = 1
		ue.Progress = sql.NullString{String: util.JsonEndCode(progress), Valid: true}
		ue.ReceiveDate = mysql.NullTime{Time: time.Now(), Valid: true}

		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			c.DB.Save(&ue)
			wg.Done()
		}()

		wg.Add(1)
		go user.UpdateGifts(gifts, logtype.GIFT_EVENT, ue.EventId, &wg)

		wg.Add(1)
		go ue.SaveLog(user, 0, uint(typeGift), gifts, &wg)
		wg.Wait()

		c.DataResponse = iris.Map{"code": 1, "data": iris.Map{constants.GIFT: gifts}}
	}
}

// /event/info/payment/accumulate
func (c *EventController) GetInfoPaymentAccumulate() {
	user := c.User
	ue := models.HfUserEvent{}

	endTime := ue.EndTimeOpenDays(user, 7)
	if endTime <= 0 {
		c.DataResponse = iris.Map{"code": -1, "msg": "Event Expired"}
		return
	}

	ue = ue.GetPaymentAccumulate(user)

	progress := map[uint]interface{}{}
	util.JsonDecodeObject(ue.Progress.String, &progress)

	gift_cf := ue.GetPaymentAccumulateGift()

	for g, val := range progress {
		progress[g] = iris.Map{"status": util.StatusGift(val), constants.GIFT: gift_cf[g]}
	}

	c.DataResponse = iris.Map{"code": 1, "data": iris.Map{
		"progress": progress,
		"payment":  ue.Turn,
		"end_time": endTime,
	}}
}

// /event/gift/payment/accumulate
func (c *EventController) PostGiftPaymentAccumulate(form formValue, ctx iris.Context) {
	if c.validToken(ctx) {
		user := c.User
		ue := models.HfUserEvent{}

		endTime := ue.EndTimeOpenDays(user, 7)
		if endTime <= 0 {
			c.DataResponse = iris.Map{"code": -1, "msg": "Event Expired"}
			return
		}

		ue = ue.GetPaymentAccumulate(user)

		typeGift := uint(util.ToInt(form("type")))

		progress := map[uint]int{}
		util.JsonDecodeObject(ue.Progress.String, &progress)

		if val, ok := progress[typeGift]; !ok || val != 0 {
			c.DataResponse = iris.Map{"code": -1, "msg": "Gift invalid"}
			return
		}

		gift_cf := ue.GetPaymentAccumulateGift()

		gifts := util.InterfaceToMap(gift_cf[typeGift])
		progress[typeGift] = 1
		ue.Progress = sql.NullString{String: util.JsonEndCode(progress), Valid: true}
		ue.ReceiveDate = mysql.NullTime{Time: time.Now(), Valid: true}

		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			c.DB.Save(&ue)
			wg.Done()
		}()

		wg.Add(1)
		go user.UpdateGifts(gifts, logtype.GIFT_EVENT, ue.EventId, &wg)

		wg.Add(1)
		go ue.SaveLog(user, 0, uint(typeGift), gifts, &wg)
		wg.Wait()

		c.DataResponse = iris.Map{"code": 1, "data": iris.Map{constants.GIFT: gifts}}
	}
}

// /event/info/payment/accumulate/v2
func (c *EventController) GetInfoPaymentAccumulateV2() {
	user := c.User
	ue := models.HfUserEvent{}

	ue, event, check := ue.GetPaymentAccumulateV2(user)
	if !check {
		c.DataResponse = iris.Map{"code": -1, "msg": "Event Expired"}
		return
	}

	progress := map[uint]interface{}{}
	util.JsonDecodeObject(ue.Progress.String, &progress)

	gift_cf := ue.GetPaymentAccumulateGift()

	for g, val := range progress {
		progress[g] = iris.Map{"status": util.StatusGift(val), constants.GIFT: gift_cf[g]}
	}

	c.DataResponse = iris.Map{"code": 1, "data": iris.Map{
		"progress": progress,
		"payment":  ue.Turn,
		"end_time": event.EndDate.Time.Unix() - time.Now().Unix(),
	}}
}

// /event/gift/payment/accumulate/v2
func (c *EventController) PostGiftPaymentAccumulateV2(form formValue, ctx iris.Context) {
	if c.validToken(ctx) {
		user := c.User
		ue := models.HfUserEvent{}

		ue, _, check := ue.GetPaymentAccumulateV2(user)
		if !check {
			c.DataResponse = iris.Map{"code": -1, "msg": "Event Expired"}
			return
		}

		typeGift := uint(util.ToInt(form("type")))

		progress := map[uint]int{}
		util.JsonDecodeObject(ue.Progress.String, &progress)

		if val, ok := progress[typeGift]; !ok || val != 0 {
			c.DataResponse = iris.Map{"code": -1, "msg": "Gift invalid"}
			return
		}

		gift_cf := ue.GetPaymentAccumulateGift()

		gifts := util.InterfaceToMap(gift_cf[typeGift])
		progress[typeGift] = 1
		ue.Progress = sql.NullString{String: util.JsonEndCode(progress), Valid: true}
		ue.ReceiveDate = mysql.NullTime{Time: time.Now(), Valid: true}

		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			c.DB.Save(&ue)
			wg.Done()
		}()

		wg.Add(1)
		go user.UpdateGifts(gifts, logtype.GIFT_EVENT, ue.EventId, &wg)

		wg.Add(1)
		go ue.SaveLog(user, 0, uint(typeGift), gifts, &wg)
		wg.Wait()

		c.DataResponse = iris.Map{"code": 1, "data": iris.Map{constants.GIFT: gifts}}
	}
}

// /event/info/turn/basic
func (c *EventController) GetInfoTurnBasic() {
	user := c.User
	ue := models.HfUserEvent{}

	endTime := ue.EndTimeOpenDays(user, 7)
	if endTime <= 0 {
		c.DataResponse = iris.Map{"code": -1, "msg": "Event Expired"}
		return
	}

	ue = ue.GetTurnBasic(user)

	progress := map[uint]interface{}{}
	util.JsonDecodeObject(ue.Progress.String, &progress)

	gift_bonus := ue.GetTurnBasicBonus()

	for g, val := range progress {
		progress[g] = iris.Map{"status": util.StatusGift(val), constants.GIFT: gift_bonus[g]}
	}

	gift_turn := ue.GetTurnBasicGift()

	c.DataResponse = iris.Map{"code": 1, "data": iris.Map{
		"turn":        ue.Turn,
		"turn_fee_1":  50,
		"turn_fee_10": 450,
		"progress":    progress,
		"gift_turn":   gift_turn,
		"end_time":    endTime,
	}}
}

// /event/gift/turn/basic
func (c *EventController) PostGiftTurnBasic(form formValue, ctx iris.Context) {
	if c.validToken(ctx) {
		user := c.User
		ue := models.HfUserEvent{}

		endTime := ue.EndTimeOpenDays(user, 7)
		if endTime <= 0 {
			c.DataResponse = iris.Map{"code": -1, "msg": "Event Expired"}
			return
		}

		ue = ue.GetTurnBasic(user)

		typeGift := util.ToInt(form("type"))
		if !util.InArray(typeGift, []int{1, 10}) {
			c.DataResponse = iris.Map{"code": -1, "msg": "Type invalid"}
			return
		}

		gemFee := 50
		if typeGift == 10 {
			gemFee = 450
		}

		if user.Gem < uint(gemFee) {
			c.DataResponse = iris.Map{"code": -1, "msg": "Gem not enough"}
			return
		}

		gift_turn := ue.GetTurnBasicGift()

		giftReturn := []iris.Map{}
		var wg sync.WaitGroup
		for i := 0; i < typeGift; i++ {
			gift, indexGift, idGift := util.RandomPercentGiftEvent(gift_turn)

			giftReturn = append(giftReturn, iris.Map{constants.INDEX: indexGift, constants.GIFT: gift})

			wg.Add(2) //nhận quà luu log
			go user.UpdateGifts(gift, logtype.GIFT_EVENT, ue.EventId, &wg)
			go ue.SaveLog(user, 0, uint(idGift), gift, &wg)
			wg.Wait()
		}

		//Trừ Gem
		wg.Add(1)
		go user.SetGem(gemFee*-1, "", uint(typeGift), logtype.BUY_EVENT, ue.EventId, &wg)

		ue.Turn += uint(typeGift)

		progress := map[uint]int{}
		util.JsonDecodeObject(ue.Progress.String, &progress)
		for i, val := range progress {
			if ue.Turn >= i && val == 2 {
				progress[i] = 0
			}
		}
		ue.Progress = sql.NullString{String: util.JsonEndCode(progress), Valid: true}

		wg.Add(1) //cập nhật bonus
		go func() {
			c.DB.Save(&ue)
			wg.Done()
		}()
		wg.Wait()

		c.DataResponse = iris.Map{"code": 1, "data": iris.Map{constants.GIFT: giftReturn}}
	}
}

// /event/bonus/turn/basic
func (c *EventController) PostBonusTurnBasic(form formValue, ctx iris.Context) {
	if c.validToken(ctx) {
		user := c.User
		ue := models.HfUserEvent{}

		endTime := ue.EndTimeOpenDays(user, 7)
		if endTime <= 0 {
			c.DataResponse = iris.Map{"code": -1, "msg": "Event Expired"}
			return
		}

		ue = ue.GetTurnBasic(user)

		progress := map[uint]int{}
		util.JsonDecodeObject(ue.Progress.String, &progress)

		typeGift := uint(util.ToInt(form("type")))
		if val, ok := progress[typeGift]; !ok || val != 0 {
			c.DataResponse = iris.Map{"code": -1, "msg": "Type invalid"}
			return
		}

		gift_bonus := ue.GetTurnBasicBonus()

		gift := util.InterfaceToMap(gift_bonus[typeGift])

		var wg sync.WaitGroup
		wg.Add(3) //nhận quà luu log đánh dấu
		go user.UpdateGifts(gift, logtype.GIFT_EVENT, ue.EventId, &wg)
		go ue.SaveLog(user, 0, typeGift, gift, &wg)
		progress[typeGift] = 1
		ue.Progress = sql.NullString{String: util.JsonEndCode(progress), Valid: true}
		go func() {
			c.DB.Save(&ue)
			wg.Done()
		}()
		wg.Wait()

		c.DataResponse = iris.Map{"code": 1, "data": iris.Map{constants.GIFT: gift}}
	}
}

// /event/info/turn/basic/v2
func (c *EventController) GetInfoTurnBasicV2() {
	user := c.User
	ue := models.HfUserEvent{}

	ue, event, check := ue.GetTurnBasicV2(user)
	if !check {
		c.DataResponse = iris.Map{"code": -1, "msg": "Event Expired"}
		return
	}

	progress := map[uint]interface{}{}
	util.JsonDecodeObject(ue.Progress.String, &progress)

	gift_bonus := ue.GetTurnBasicBonus()

	for g, val := range progress {
		progress[g] = iris.Map{"status": util.StatusGift(val), constants.GIFT: gift_bonus[g]}
	}

	gift_turn := ue.GetTurnBasicGift()

	c.DataResponse = iris.Map{"code": 1, "data": iris.Map{
		"turn":        ue.Turn,
		"turn_fee_1":  50,
		"turn_fee_10": 450,
		"progress":    progress,
		"gift_turn":   gift_turn,
		"end_time":    event.EndDate.Time.Unix() - time.Now().Unix(),
	}}
}

// /event/gift/turn/basic/v2
func (c *EventController) PostGiftTurnBasicV2(form formValue, ctx iris.Context) {
	if c.validToken(ctx) {
		user := c.User
		ue := models.HfUserEvent{}

		ue, _, check := ue.GetTurnBasicV2(user)
		if !check {
			c.DataResponse = iris.Map{"code": -1, "msg": "Event Expired"}
			return
		}

		typeGift := util.ToInt(form("type"))
		if !util.InArray(typeGift, []int{1, 10}) {
			c.DataResponse = iris.Map{"code": -1, "msg": "Type invalid"}
			return
		}

		gemFee := 50
		if typeGift == 10 {
			gemFee = 450
		}

		if user.Gem < uint(gemFee) {
			c.DataResponse = iris.Map{"code": -1, "msg": "Gem not enough"}
			return
		}

		gift_turn := ue.GetTurnBasicGift()

		giftReturn := []iris.Map{}
		var wg sync.WaitGroup
		for i := 0; i < typeGift; i++ {
			gift, indexGift, idGift := util.RandomPercentGiftEvent(gift_turn)

			giftReturn = append(giftReturn, iris.Map{constants.INDEX: indexGift, constants.GIFT: gift})

			wg.Add(2) //nhận quà luu log
			go user.UpdateGifts(gift, logtype.GIFT_EVENT, ue.EventId, &wg)
			go ue.SaveLog(user, 0, uint(idGift), gift, &wg)
			wg.Wait()
		}

		//Trừ Gem
		wg.Add(1)
		go user.SetGem(gemFee*-1, "", uint(typeGift), logtype.BUY_EVENT, ue.EventId, &wg)

		ue.Turn += uint(typeGift)

		progress := map[uint]int{}
		util.JsonDecodeObject(ue.Progress.String, &progress)
		for i, val := range progress {
			if ue.Turn >= i && val == 2 {
				progress[i] = 0
			}
		}
		ue.Progress = sql.NullString{String: util.JsonEndCode(progress), Valid: true}

		wg.Add(1) //cập nhật bonus
		go func() {
			c.DB.Save(&ue)
			wg.Done()
		}()
		wg.Wait()

		c.DataResponse = iris.Map{"code": 1, "data": iris.Map{constants.GIFT: giftReturn}}
	}
}

// /event/bonus/turn/basic/v2
func (c *EventController) PostBonusTurnBasicV2(form formValue, ctx iris.Context) {
	if c.validToken(ctx) {
		user := c.User
		ue := models.HfUserEvent{}

		ue, _, check := ue.GetTurnBasicV2(user)
		if !check {
			c.DataResponse = iris.Map{"code": -1, "msg": "Event Expired"}
			return
		}

		progress := map[uint]int{}
		util.JsonDecodeObject(ue.Progress.String, &progress)

		typeGift := uint(util.ToInt(form("type")))
		if val, ok := progress[typeGift]; !ok || val != 0 {
			c.DataResponse = iris.Map{"code": -1, "msg": "Type invalid"}
			return
		}

		gift_bonus := ue.GetTurnBasicBonus()

		gift := util.InterfaceToMap(gift_bonus[typeGift])

		var wg sync.WaitGroup
		wg.Add(3) //nhận quà luu log đánh dấu
		go user.UpdateGifts(gift, logtype.GIFT_EVENT, ue.EventId, &wg)
		go ue.SaveLog(user, 0, typeGift, gift, &wg)
		progress[typeGift] = 1
		ue.Progress = sql.NullString{String: util.JsonEndCode(progress), Valid: true}
		go func() {
			c.DB.Save(&ue)
			wg.Done()
		}()
		wg.Wait()

		c.DataResponse = iris.Map{"code": 1, "data": iris.Map{constants.GIFT: gift}}
	}
}

// /event/info/gift/discount
func (c *EventController) GetInfoGiftDiscount() {
	user := c.User
	ue := models.HfUserEvent{}

	endTime := ue.EndTimeOpenDays(user, 4)
	if endTime <= 0 {
		c.DataResponse = iris.Map{"code": -1, "msg": "Event Expired"}
		return
	}

	ue = ue.GetGiftDiscount(user)

	progress := map[int]interface{}{}
	util.JsonDecodeObject(ue.Progress.String, &progress)

	gift_cf := ue.GetGiftDiscountGift()

	for idGift, item := range gift_cf {
		item[constants.BOUGHT] = progress[idGift]
		gift_cf[idGift] = item
	}

	c.DataResponse = iris.Map{"code": 1, "data": iris.Map{
		"gift_shop": gift_cf,
		"end_time":  endTime,
	}}
}

// /event/gift/discount
func (c *EventController) PostGiftDiscount(form formValue, ctx iris.Context) {
	if c.validToken(ctx) {
		user := c.User
		ue := models.HfUserEvent{}

		endTime := ue.EndTimeOpenDays(user, 4)
		if endTime <= 0 {
			c.DataResponse = iris.Map{"code": -1, "msg": "Event Expired"}
			return
		}

		ue = ue.GetGiftDiscount(user)

		progress := map[int]int{}
		util.JsonDecodeObject(ue.Progress.String, &progress)

		idGift := util.ToInt(form("id"))

		gift_cf := ue.GetGiftDiscountGift()

		if bought, ok := progress[idGift]; !ok || bought >= util.ToInt(gift_cf[idGift][constants.LIMIT]) {
			c.DataResponse = iris.Map{"code": -1, "msg": "Id gift invalid or limit bought"}
			return
		}

		gemFee := util.ToInt(gift_cf[idGift][constants.PRICE])
		if user.Gem < uint(gemFee) {
			c.DataResponse = iris.Map{"code": -1, "msg": "Gem not enough"}
			return
		}

		gift := util.InterfaceToMap(gift_cf[idGift][constants.GIFT])

		var wg sync.WaitGroup
		wg.Add(4) //trừ gen nhận quà luu log đánh dấu
		go user.SetGem(gemFee*-1, "", uint(idGift), logtype.BUY_EVENT, ue.EventId, &wg)
		go user.UpdateGifts(gift, logtype.GIFT_EVENT, ue.EventId, &wg)
		go ue.SaveLog(user, 0, uint(idGift), gift, &wg)

		progress[idGift]++
		ue.Progress = sql.NullString{String: util.JsonEndCode(progress), Valid: true}
		go func() {
			c.DB.Save(&ue)
			wg.Done()
		}()
		wg.Wait()

		c.DataResponse = iris.Map{"code": 1, "data": iris.Map{constants.GIFT: gift}}
	}
}

// /event/info/gift/discount/v2
func (c *EventController) GetInfoGiftDiscountV2() {
	user := c.User
	ue := models.HfUserEvent{}

	ue, event, check := ue.GetGiftDiscountV2(user)

	if !check  {
		c.DataResponse = iris.Map{"code": -1, "msg": "Event Expired"}
		return
	}

	progress := map[int]interface{}{}
	util.JsonDecodeObject(ue.Progress.String, &progress)

	gift_cf := ue.GetGiftDiscountGift()

	for idGift, item := range gift_cf {
		item[constants.BOUGHT] = progress[idGift]
		gift_cf[idGift] = item
	}

	c.DataResponse = iris.Map{"code": 1, "data": iris.Map{
		"gift_shop": gift_cf,
		"end_time":  event.EndDate.Time.Unix() - time.Now().Unix(),
	}}
}

// /event/gift/discount/v2
func (c *EventController) PostGiftDiscountV2(form formValue, ctx iris.Context) {
	if c.validToken(ctx) {
		user := c.User
		ue := models.HfUserEvent{}

		ue, _, check := ue.GetGiftDiscountV2(user)

		if !check {
			c.DataResponse = iris.Map{"code": -1, "msg": "Event Expired"}
			return
		}

		progress := map[int]int{}
		util.JsonDecodeObject(ue.Progress.String, &progress)

		idGift := util.ToInt(form("id"))

		gift_cf := ue.GetGiftDiscountGift()

		if bought, ok := progress[idGift]; !ok || bought >= util.ToInt(gift_cf[idGift][constants.LIMIT]) {
			c.DataResponse = iris.Map{"code": -1, "msg": "Id gift invalid or limit bought"}
			return
		}

		gemFee := util.ToInt(gift_cf[idGift][constants.PRICE])
		if user.Gem < uint(gemFee) {
			c.DataResponse = iris.Map{"code": -1, "msg": "Gem not enough"}
			return
		}

		gift := util.InterfaceToMap(gift_cf[idGift][constants.GIFT])

		var wg sync.WaitGroup
		wg.Add(4) //trừ gen nhận quà luu log đánh dấu
		go user.SetGem(gemFee*-1, "", uint(idGift), logtype.BUY_EVENT, ue.EventId, &wg)
		go user.UpdateGifts(gift, logtype.GIFT_EVENT, ue.EventId, &wg)
		go ue.SaveLog(user, 0, uint(idGift), gift, &wg)

		progress[idGift]++
		ue.Progress = sql.NullString{String: util.JsonEndCode(progress), Valid: true}
		go func() {
			c.DB.Save(&ue)
			wg.Done()
		}()
		wg.Wait()

		c.DataResponse = iris.Map{"code": 1, "data": iris.Map{constants.GIFT: gift}}
	}
}

// /event/info/gift/time
func (c *EventController) GetInfoGiftTime() {
	user := c.User
	ue := models.HfUserEvent{}

	ue, ev, check := ue.GetGiftTime(user)

	if !check {
		c.DataResponse = iris.Map{"code": -1, "msg": "Event Expired"}
		return
	}

	giftCf := map[uint]iris.Map{}
	util.JsonDecodeObject(ev.Config, &giftCf)

	if _, ok := giftCf[ue.Turn]; !ok {
		c.DataResponse = iris.Map{"code": -1, "msg": "Out of gifts"}
		return
	}

	gift := giftCf[ue.Turn][constants.GIFT]
	timeCf := util.ToInt(giftCf[ue.Turn][constants.TIME])

	timeGift := ue.ReceiveDate.Time.Add(time.Second * time.Duration(timeCf)).Unix() - time.Now().Unix()
	if timeGift < 0 {
		timeGift = 0
	}

	c.DataResponse = iris.Map{
		"code": 1,
		"data": iris.Map{
			"next_gift": gift,
			constants.TIME: timeGift,
		},
	}
}

// /event/gift/time
func (c *EventController) PostGiftTime(form formValue, ctx iris.Context) {
	if c.validToken(ctx) {
		user := c.User
		ue := models.HfUserEvent{}

		ue, ev, check := ue.GetGiftTime(user)

		if !check {
			c.DataResponse = iris.Map{"code": -1, "msg": "Event Expired"}
			return
		}

		giftCf := map[uint]iris.Map{}
		util.JsonDecodeObject(ev.Config, &giftCf)

		if _, ok := giftCf[ue.Turn]; !ok {
			c.DataResponse = iris.Map{"code": -1, "msg": "Out of gifts"}
			return
		}
		gift := util.InterfaceToMap(giftCf[ue.Turn][constants.GIFT])
		timeCf := util.ToInt(giftCf[ue.Turn][constants.TIME])

		timeGift := ue.ReceiveDate.Time.Add(time.Second * time.Duration(timeCf)).Unix() - time.Now().Unix()
		if timeGift > 0 {
			c.DataResponse = iris.Map{"code": -1, "msg": "Please wait"}
			return
		}

		ue.Turn++
		nextGift, timeCfNext := iris.Map{}, 0
		nextGift = nil
		if val, ok := giftCf[ue.Turn]; ok {
			nextGift = util.InterfaceToMap(val[constants.GIFT])
			timeCfNext = util.ToInt(val[constants.TIME])
		}

		ue.ReceiveDate = mysql.NullTime{Valid: true, Time: time.Now()}

		var wg sync.WaitGroup
		wg.Add(3)
		go user.UpdateGifts(gift, logtype.GIFT_TIME, ue.EventId, &wg)
		go ue.SaveLog(user, 0, ue.Turn-1, gift, &wg)
		go func() {
			c.DB.Save(&ue)
			wg.Done()
		}()
		wg.Wait()

		c.DataResponse = iris.Map{
			"code": 1,
			"data": iris.Map{
				constants.GIFT: gift,
				"next_gift":    nextGift,
				constants.TIME: timeCfNext,
			},
		}
	}
}

