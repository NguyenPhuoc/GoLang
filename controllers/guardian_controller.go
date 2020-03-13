package controllers

import (
	"GoLang/libraries/constants"
	"GoLang/libraries/logtype"
	"GoLang/libraries/util"
	"GoLang/models"
	"database/sql"
	"fmt"
	"github.com/jinzhu/now"
	"github.com/kataras/iris"
	"math"
	"sync"
	"time"
)

type GuardianController struct {
	MyController
}

// /guardian/info
func (c *GuardianController) GetInfo() {
	user := c.User

	uhg := models.HfUserHuntGuardian{}
	data := uhg.GetMaps(user)

	guardianTicket := 10
	ugi := models.HfUserGuardianInfo{}
	ugi, check := ugi.Get(user)
	if check {
		guardianTicket = int(ugi.Ticket)
	} else {
		var wg sync.WaitGroup
		wg.Add(1)
		go ugi.SetTicket(guardianTicket, "", 0, logtype.GET_INFO_TICKET_GUARDIAN, 0, user, &wg)
		wg.Wait()
	}

	timeReset := now.EndOfWeek().Unix() - time.Now().Unix()
	gemPriceConfig := util.ToInt(user.GetConfig("guardian_ticket_price").Value)

	c.IsEncrypt = true
	c.DataResponse = iris.Map{"code": 1, "data": iris.Map{
		constants.TICKET_GUARDIAN: guardianTicket,
		"guardian_hunt":           data,
		"time_reset":              timeReset,
		"price_ticket":            gemPriceConfig,
	}}
}

// /guardian/buy/ticket
func (c *GuardianController) PostBuyTicket(form formValue, ctx iris.Context) {
	if c.validToken(ctx) {
		user := c.User

		gemPriceConfig := util.ToInt(user.GetConfig("guardian_ticket_price").Value)
		quantity := math.Max(float64(util.ParseInt(form("quantity"))), 1)

		if user.Gem < uint(gemPriceConfig)*uint(quantity) {
			c.DataResponse = iris.Map{"code": -1, "msg": "Not enough Gem"}
			return
		}

		//trừ gem + item
		ugi := models.HfUserGuardianInfo{}
		ugi.Get(user)

		var wg sync.WaitGroup
		wg.Add(1)
		go ugi.SetTicket(int(quantity), "", 0, logtype.BUY_TICKET_GUARDIAN, 0, user, &wg)

		wg.Add(1)
		gemFee := (gemPriceConfig * int(quantity)) * -1
		go user.SetGem(gemFee, "", 0, logtype.BUY_TICKET_GUARDIAN, 0, &wg)
		wg.Wait()

		c.DataResponse = iris.Map{"code": 1, "data": iris.Map{constants.GIFT: iris.Map{constants.TICKET_GUARDIAN: quantity}, constants.GEM: user.Gem}}
	}
}

// /guardian/starthunt
func (c *GuardianController) PostStarthunt(form formValue, ctx iris.Context) {
	if c.validToken(ctx) {
		user := c.User

		ugi := models.HfUserGuardianInfo{}
		ugi.Get(user)

		if ugi.Ticket < 1 {
			c.DataResponse = iris.Map{"code": -1, "msg": "Not enough ticket"}
			return
		}

		var wg sync.WaitGroup
		wg.Add(1)
		go ugi.SetTicket(-1, "", 0, logtype.STARTGAME_TICKET_GUARDIAN, 0, user, &wg)
		wg.Wait()

		c.DataResponse = iris.Map{"code": 1, "msg": "Success"}
	}
}

// /guardian/endhunt
func (c *GuardianController) PostEndhunt(form formValue, ctx iris.Context) {
	if c.validToken(ctx) {
		user := c.User
		hp := util.ParseInt(form("hp"))
		guardianId := uint16(util.ParseInt(form("guardian_id")))
		data := form("data")
		level := util.ParseInt(form("level"))
		lastLineUp := form("last_line_up")
		powerPoint := util.ParseInt(form("power_point"))

		gua := models.HfGuardianHunt{}
		gua, checkGua := gua.Find(guardianId, 1, user)
		if !checkGua {
			c.DataResponse = iris.Map{"code": -1, "msg": "Invalid guardianId"}
			return
		}
		//todo hp.ToString() + guardianId.ToString() + level.ToString() + powerPoint.ToString() + lastLineUp
		if !util.ValidHash(form, hp, guardianId, level, powerPoint, lastLineUp) {
			c.DataResponse = iris.Map{"code": -1, "msg": "Invalid hash"}
			return
		}

		uHuntGuardian := models.HfUserHuntGuardian{}
		uHuntGuardian = uHuntGuardian.Get(guardianId, user)
		_, checkGua = gua.Find(guardianId, uHuntGuardian.GuardianLevel, user)
		if !checkGua {
			c.DataResponse = iris.Map{"code": -1, "msg": "User hack"}
			return
		}

		//cập nhật level, powerPoint, lastLineUp
		ugi := models.HfUserGuardianInfo{}
		ugi.Get(user)
		ugi.Level = level
		ugi.PowerPoint = powerPoint
		ugi.LastLineUp = sql.NullString{String: lastLineUp, Valid: true}
		c.DB.Save(&ugi)

		//trả quà đánh
		guaAward := models.HfGuardianAward{}
		guaAward.Find(uHuntGuardian.GuardianLevel, user)
		gifts := util.JsonDecodeMap(guaAward.GiftHunt)

		//trừ máu đánh được
		uHuntGuardian.GuardianHp -= hp

		winGua := false
		//lên cấp nếu đánh thắng
		guardianLevel := uHuntGuardian.GuardianLevel
		if uHuntGuardian.GuardianHp <= 0 {
			uHuntGuardian.GuardianLevel += 1
			gua, checkGua := gua.Find(uHuntGuardian.GuardianId, uHuntGuardian.GuardianLevel, user)
			if checkGua {
				uHuntGuardian.GuardianHp = gua.Hp
			}
			winGua = true
		}

		var wg sync.WaitGroup
		if winGua {
			gua := models.HfGuardian{}
			gua.Find(uHuntGuardian.GuardianId, user)

			guaAward := models.HfGuardianAward{}
			guaAward.Find(uHuntGuardian.GuardianLevel-1, user)
			gifts := util.JsonDecodeMap(guaAward.GiftKill)
			gifts[constants.PIECE_GUARDIAN] = map[uint16]int{uHuntGuardian.GuardianId: guaAward.Piece}

			ui := models.HfUserInbox{}
			ui.Id = util.UUID()
			ui.ReceiverId = user.UserId
			ui.SenderType = constants.INBOX_SENDER_BY_EVENT
			ui.TypeLog = logtype.KILL_GUARDIAN
			ui.IsReceive = 0

			title := fmt.Sprintf("Quà chiến thắng hộ vệ %s cấp %d", gua.Name, guaAward.Level)

			ui.Title = title
			ui.Gift = sql.NullString{Valid: true, String: util.JsonEndCode(gifts)}
			ui.CreatedDate = time.Now()
			wg.Add(1)
			go func() {
				c.DB.Save(&ui)
				wg.Done()
			}()
		}

		wg.Add(1)
		go user.UpdateGifts(gifts, logtype.ENDGAME_GUARDIAN, 0, &wg)

		wg.Add(1)
		go func() {
			c.DB.Save(&uHuntGuardian)
			wg.Done()
		}()

		wg.Add(1)
		dataLog := iris.Map{"server_id": user.ServerId, "user_id": user.UserId, "guardian_id": guardianId, "guardian_level": guardianLevel, "hp": hp, "data": util.JsonDecodeMap(data), "is_win": winGua, "created_date": time.Now()}
		go user.SaveLogMongo(logtype.HF_USER_HUNT_GUARDIAN_LOG, dataLog, &wg)

		wg.Add(1)
		go user.CompleteMissionDaily(models.QDL_HUNTGUARDIAN, &wg)
		wg.Wait()

		c.DataResponse = iris.Map{
			"code": 1,
			"msg":  "Success",
			"data": iris.Map{
				constants.GIFT: gifts,
			},
		}
	}
}

// /guardian/combine
func (c *GuardianController) PostCombine(form formValue, ctx iris.Context) {
	c.IsEncrypt = true

	if c.validToken(ctx) {
		user := c.User

		guardianId := uint16(util.ToInt(form("guardian_id")))

		gu := models.HfGuardianUpgrade{}
		gu, check := gu.Find(guardianId, 0, user)
		if !check {
			c.DataResponse = iris.Map{"code": -1, "msg": "Guardian not found"}
			return
		}

		uGua := models.HfUserGuardian{}
		uGua, check = uGua.Find(guardianId, user)
		if uGua.Stat != 0 {
			c.DataResponse = iris.Map{"code": -1, "msg": "Guardian has owned"}
			return
		}

		if !check || uGua.Piece < gu.Piece {
			c.DataResponse = iris.Map{"code": -1, "msg": "Piece not enough"}
			return
		}

		var wg sync.WaitGroup
		wg.Add(1)

		uGua.Stat = 1
		uGua.Level = 1
		uGua.LevelPassive1 = 1

		pieceFee := int(gu.Piece) * -1
		go uGua.SetPiece(pieceFee, 0, logtype.COMBINE_GUARDIAN, 0, user, &wg)
		wg.Wait()

		c.DataResponse = iris.Map{"code": 1, "data": uGua.GetMap()}
	}
}

// /guardian/upgrade/level
func (c *GuardianController) PostUpgradeLevel(form formValue, ctx iris.Context) {
	c.IsEncrypt = true

	if c.validToken(ctx) {
		user := c.User

		guardianId := uint16(util.ToInt(form("guardian_id")))

		gu := models.HfGuardianUpgrade{}
		gu, check := gu.Find(guardianId, 0, user)
		if !check {
			c.DataResponse = iris.Map{"code": -1, "msg": "Guardian not found"}
			return
		}

		uGua := models.HfUserGuardian{}
		uGua, check = uGua.Find(guardianId, user)
		if uGua.Stat != 1 {
			c.DataResponse = iris.Map{"code": -1, "msg": "Guardian not own"}
			return
		}

		if uGua.Level >= (uGua.Evolve+1)*30 {
			c.DataResponse = iris.Map{"code": -1, "msg": "Level's Max"}
			return
		}

		uGuaInfo := models.HfUserGuardianInfo{}
		uGuaInfo.Get(user)

		nextLevel := uGua.Level + 1
		fruitRequest := util.ToInt(util.EvaluateMath(gu.FruitFormula, map[string]interface{}{constants.LEVEL: nextLevel}))
		goldRequest := util.ToInt(util.EvaluateMath(gu.GoldFormula, map[string]interface{}{constants.LEVEL: nextLevel}))

		if user.Gold < uint(goldRequest) || uGuaInfo.Fruit < uint(fruitRequest) {
			c.DataResponse = iris.Map{"code": -1, "msg": "Gold:" + util.ToString(goldRequest) + " or Fruit:" + util.ToString(fruitRequest) + " not enough"}
			return
		}

		var wg sync.WaitGroup
		wg.Add(2)
		go user.SetGold(goldRequest*-1, uGua.GuardianId, uint(nextLevel), logtype.UPDRADE_LEVEL_GUARDIAN, 0, &wg)
		go uGuaInfo.SetFruit(fruitRequest*-1, uGua.GuardianId, uint(nextLevel), logtype.UPDRADE_LEVEL_GUARDIAN, 0, user, &wg)

		wg.Add(1)
		uGua.Level = nextLevel
		go func() {
			c.DB.Save(&uGua)
			wg.Done()
		}()
		wg.Wait()

		c.DataResponse = iris.Map{"code": 1, "data": uGua.GetMap()}
	}
}

// /guardian/evolve
func (c *GuardianController) PostEvolve(form formValue, ctx iris.Context) {
	c.IsEncrypt = true

	if c.validToken(ctx) {
		user := c.User

		guardianId := uint16(util.ToInt(form("guardian_id")))

		uGua := models.HfUserGuardian{}
		uGua, check := uGua.Find(guardianId, user)
		if !check || uGua.Stat != 1 {
			c.DataResponse = iris.Map{"code": -1, "msg": "Guardian not own"}
			return
		}

		if uGua.Evolve >= 3 {
			c.DataResponse = iris.Map{"code": -1, "msg": "Evolve's Max"}
			return
		}

		if uGua.Level < (uGua.Evolve+1)*30 {
			c.DataResponse = iris.Map{"code": -1, "msg": "Level not enough"}
			return
		}

		uGuaInfo := models.HfUserGuardianInfo{}
		uGuaInfo.Get(user)

		nextEvolve := uGua.Evolve + 1

		gu := models.HfGuardianUpgrade{}
		gu, check = gu.Find(guardianId, nextEvolve, user)
		if !check {
			c.DataResponse = iris.Map{"code": -1, "msg": "Guardian not found"}
			return
		}

		fruitRequest := util.ToInt(util.EvaluateMath(gu.FruitFormula, map[string]interface{}{}))
		goldRequest := util.ToInt(util.EvaluateMath(gu.GoldFormula, map[string]interface{}{}))
		pieceRequest := util.ToInt(gu.Piece)

		if user.Gold < uint(goldRequest) || uGuaInfo.Fruit < uint(fruitRequest) || uGua.Piece < uint(pieceRequest) {
			c.DataResponse = iris.Map{"code": -1, "msg": "Gold:" + util.ToString(goldRequest) + " or Fruit:" + util.ToString(fruitRequest) + " or Piece:" + util.ToString(pieceRequest) + " not enough"}
			return
		}

		var wg sync.WaitGroup
		wg.Add(3)
		go user.SetGold(goldRequest*-1, uGua.GuardianId, uint(nextEvolve), logtype.EVOLVE_GUARDIAN, 0, &wg)
		go uGuaInfo.SetFruit(fruitRequest*-1, uGua.GuardianId, uint(nextEvolve), logtype.EVOLVE_GUARDIAN, 0, user, &wg)

		uGua.Evolve = nextEvolve
		switch nextEvolve {
		case 1:
			uGua.LevelPassive2 = 1
		case 2:
			uGua.LevelPassive3 = 1
		case 3:
			uGua.LevelPassive4 = 1
		}
		go uGua.SetPiece(pieceRequest*-1, uint(nextEvolve), logtype.EVOLVE_GUARDIAN, 0, user, &wg)
		wg.Wait()

		c.DataResponse = iris.Map{"code": 1, "data": uGua.GetMap()}
	}
}

// /guardian/passive/level
func (c *GuardianController) PostPassiveLevel(form formValue, ctx iris.Context) {
	c.IsEncrypt = true

	if c.validToken(ctx) {
		user := c.User

		guardianId := uint16(util.ToInt(form("guardian_id")))
		passiveId := uint16(util.ToInt(form("passive_id")))

		uGua := models.HfUserGuardian{}
		uGua, check := uGua.Find(guardianId, user)
		if !check || uGua.Stat != 1 {
			c.DataResponse = iris.Map{"code": -1, "msg": "Guardian not own"}
			return
		}

		if uGua.Evolve+1 < uint16(passiveId) {
			c.DataResponse = iris.Map{"code": -1, "msg": "Passive not active"}
			return
		}

		passiveLevel := uint16(0)
		switch passiveId {
		case 1:
			passiveLevel = uGua.LevelPassive1
		case 2:
			passiveLevel = uGua.LevelPassive2
		case 3:
			passiveLevel = uGua.LevelPassive3
		case 4:
			passiveLevel = uGua.LevelPassive4
		default:
			c.DataResponse = iris.Map{"code": -1, "msg": "passive id invalid"}
			return
		}

		passiveNextLevel := passiveLevel + 1
		if passiveLevel >= 30 {
			c.DataResponse = iris.Map{"code": -1, "msg": "Passive Level's Max"}
			return
		}

		uGuaInfo := models.HfUserGuardianInfo{}
		uGuaInfo.Get(user)

		gu := models.HfGuardianUpgrade{}
		gu, check = gu.Find(guardianId, passiveId-1, user)
		if !check {
			c.DataResponse = iris.Map{"code": -1, "msg": "Guardian not found"}
			return
		}

		flowerRequest := util.ToInt(util.EvaluateMath(gu.FlowerFormulaPassive, map[string]interface{}{constants.LEVEL: passiveNextLevel}))
		goldRequest := util.ToInt(util.EvaluateMath(gu.GoldFormulaPassive, map[string]interface{}{constants.LEVEL: passiveNextLevel}))

		if user.Gold < uint(goldRequest) || uGuaInfo.Flower < uint(flowerRequest) {
			c.DataResponse = iris.Map{"code": -1, "msg": "Gold:" + util.ToString(goldRequest) + " or Flower:" + util.ToString(flowerRequest) + " not enough"}
			return
		}

		var wg sync.WaitGroup
		wg.Add(3)
		go user.SetGold(goldRequest*-1, uGua.GuardianId, uint(passiveNextLevel), logtype.PASSIVE_LEVEL_GUARDIAN, 0, &wg)
		go uGuaInfo.SetFlower(flowerRequest*-1, uGua.GuardianId, uint(passiveNextLevel), logtype.PASSIVE_LEVEL_GUARDIAN, 0, user, &wg)

		switch passiveId {
		case 1:
			uGua.LevelPassive1 = passiveNextLevel
		case 2:
			uGua.LevelPassive2 = passiveNextLevel
		case 3:
			uGua.LevelPassive3 = passiveNextLevel
		case 4:
			uGua.LevelPassive4 = passiveNextLevel
		}
		go func() {
			c.DB.Save(&uGua)
			wg.Done()
		}()
		wg.Wait()

		c.DataResponse = iris.Map{"code": 1, "data": uGua.GetMap()}
	}
}
