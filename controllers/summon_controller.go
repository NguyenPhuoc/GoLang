package controllers

import (
	"GoLang/libraries/constants"
	"GoLang/libraries/util"
	"GoLang/models"
	"github.com/go-sql-driver/mysql"
	"github.com/kataras/iris"
	"math"
	"sync"
	"time"
)

type SummonController struct {
	MyController
}

// /summon/info
func (c *SummonController) GetInfo() {
	user := c.User

	uSum1 := models.HfUserSummon{}
	uSum1.Get(1, user)
	summon_1 := uSum1.GetMap(user)

	uSum2 := models.HfUserSummon{}
	uSum2.Get(2, user)
	summon_2 := uSum2.GetMap(user)

	uSum3 := models.HfUserSummon{}
	uSum3.Get(3, user)
	summon_3 := uSum3.GetMap(user)

	c.DataResponse = iris.Map{"code": 1, "data": iris.Map{
		"summon_1": summon_1,
		"summon_2": summon_2,
		"summon_3": summon_3,
	}}
}

// /summon/turn
func (c *SummonController) PostTurn(form formValue, ctx iris.Context) {
	if c.validToken(ctx) {
		user := c.User

		turnType := form("type")
		summonId := uint8(util.ToInt(form("summon_id")))
		quantity := int(math.Max(1, float64(util.ToInt(form("quantity")))))

		uSum := models.HfUserSummon{}
		uSum, checkSum := uSum.Get(summonId, user)

		sum := models.HfSummon{}
		sum.Find(summonId, user)

		if !checkSum || (quantity != 1 && quantity != 10) {
			c.DataResponse = iris.Map{"code": -1, "msg": "Summon id or quantity invalid"}
			return
		}

		checkTurn := false
		data := uSum.GetMap(user)
		gemFee := 0
		if turnType == "free" {

			quantity = 1
			if util.ToInt(data["turn_free"]) == 0 {
				checkTurn = true
			}
		} else if turnType == "turn" {

			if uint(quantity) <= uSum.Quantity {
				checkTurn = true
			}
		} else if turnType == "gem" {

			if quantity == 1 {

				gemFee = util.ToInt(data["turn_fee_1"])
				if uint(gemFee) <= user.Gem && gemFee > 0 {
					checkTurn = true
				}
			} else if quantity == 10 {

				gemFee = util.ToInt(data["turn_fee_10"])
				if uint(gemFee) <= user.Gem && gemFee > 0{
					checkTurn = true
				}
			}
		}

		if !checkTurn {
			c.DataResponse = iris.Map{"code": -1, "msg": "Turn invalid"}
			return
		}

		var wg sync.WaitGroup
		wg.Add(2)

		gifts, giftReturn := uSum.Random(quantity, user)

		logType := uSum.GetLogType(turnType)
		if turnType == "free" {
			go func() {
				uSum.TurnFree = mysql.NullTime{Time: time.Now(), Valid: true}
				c.DB.Save(&uSum)
				wg.Done()
			}()
		} else if turnType == "turn" {
			quantityFee := quantity * -1
			go uSum.SetQuantity(quantityFee, 0, logType, 0, user, &wg)
		} else if turnType == "gem" {
			gemFee *= -1
			go user.SetGem(gemFee, uSum.SummonId, uint(quantity), logType, 0, &wg)
		}

		go user.UpdateGifts(gifts, logType, 0, &wg)

		//Todo Update nhiệm vụ
		if summonId == 1 {
			wg.Add(1)
			go user.CompleteMissionDaily(models.QDL_SUMMON_1, &wg, quantity)
		} else if summonId == 2 {
			wg.Add(1)
			go user.CompleteMissionDaily(models.QDL_SUMMON_2, &wg, quantity)
		}
		wg.Wait()

		c.DataResponse = iris.Map{"code": 1, "data": iris.Map{constants.GIFT: giftReturn}}
	}
}
