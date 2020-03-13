package controllers

import (
	"GoLang/libraries/constants"
	"GoLang/libraries/logtype"
	"GoLang/libraries/util"
	"GoLang/models"
	"github.com/go-sql-driver/mysql"
	"github.com/kataras/iris"
	"sync"
	"time"
)

type RollupController struct {
	MyController
}

// /rollup/info
func (c *RollupController) GetInfo() {
	user := c.User

	uRollup := models.HfUserRollup{}
	uRollup.Get(user)

	data := uRollup.GetMap(user)

	c.DataResponse = iris.Map{"code": 1, "data": data}
}

// /rollup/checkin
func (c *RollupController) PostCheckin(form formValue, ctx iris.Context) {
	if c.validToken(ctx) {
		user := c.User

		uRollup := models.HfUserRollup{}
		uRollup.Get(user)

		uRollup.Total++
		day := util.ToString(uRollup.Total)
		progress := util.JsonDecodeMap(uRollup.Progress)

		if uRollup.TimeReceive.Valid == true && uRollup.TimeReceive.Time.Day() == time.Now().Day() {
			c.DataResponse = iris.Map{"code": -1, "msg": "has received gift"}
			return
		}

		if val, ok := progress[day]; !ok || util.ToInt(val) != 0 {
			c.DataResponse = iris.Map{"code": -1, "msg": "gift not found"}
			return
		}
		//update trang thái
		progress[day] = 1
		uRollup.Progress = util.JsonEndCode(progress)
		uRollup.TimeReceive = mysql.NullTime{Valid: true, Time: time.Now()}

		//lấy data gift
		rollup := models.HfRollup{}
		rollup.Get(uRollup.Month, user)
		giftCf := util.JsonDecodeMap(rollup.Gift)

		//nhận quà lưu log
		var wg sync.WaitGroup
		wg.Add(2)
		gifts := util.InterfaceToMap(giftCf[day])
		go user.UpdateGifts(gifts, logtype.GIFT_ROLL_UP, 0, &wg)
		go uRollup.Save(user, &wg)
		wg.Wait()

		c.DataResponse = iris.Map{"code": 1, "data": iris.Map{constants.GIFT: gifts}}
	}
}
