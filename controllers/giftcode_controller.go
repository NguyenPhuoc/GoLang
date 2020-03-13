package controllers

import (
	"GoLang/libraries/constants"
	"GoLang/libraries/logtype"
	"GoLang/libraries/util"
	"GoLang/models"
	"database/sql"
	"github.com/go-sql-driver/mysql"
	"github.com/kataras/iris"
	"github.com/kataras/iris/context"
	"strings"
	"sync"
	"time"
)

type GiftCodeController struct {
	MyController
}

// /giftcode/receive
func (c *GiftCodeController) PostReceive(ctx context.Context, value formValue) {

	if c.validToken(ctx) {
	//if true {
		user := c.User

		codeInput := strings.ToUpper(strings.TrimSpace(value("code")))

		i := strings.Index(codeInput, "_")
		codes := strings.Split(codeInput, "_")

		gc := models.HfGiftCode{}
		if i != -1 && (codes[0] == "" || codes[1] == "") {
			c.DataResponse = iris.Map{"code": -1, "msg": "gift_code_invalid"}
			gc.CheckInput(user,false)
			return
		}

		code := codes[0]
		giftCode := ""
		if i != -1 {
			giftCode = codes[1]
		}

		//chuyển qua main để check gift code
		user = c.SetDB(0, user)

		if checkInput, _ := gc.CheckInput(user, true); !checkInput {

			c.DataResponse = iris.Map{"code": -1, "msg": "gift_code_many_wrong"}
			return
		}

		gc, checkGc := gc.Find(code, user)
		if !checkGc {
			c.DataResponse = iris.Map{"code": -1, "msg": "gift_code_invalid"}
			gc.CheckInput(user,false)
			return
		}

		serverIds := util.JsonDecodeArray(gc.ServerId)
		if len(serverIds) != 0 && util.InArray(user.ServerId, serverIds) {
			c.DataResponse = iris.Map{"code": -1, "msg": "gift_code_invalid"}
			gc.CheckInput(user,false)
			return
		}

		if (gc.StartDate.Valid == true || gc.EndDate.Valid == true) &&
			(time.Now().Unix() < gc.StartDate.Time.Unix() || time.Now().Unix() > gc.EndDate.Time.Unix()) {
			c.DataResponse = iris.Map{"code": -1, "msg": "gift_code_invalid"}
			gc.CheckInput(user,false)
			return
		}

		switch gc.Type {
		case "freedom":
			gci := models.HfGiftCodeItems{}
			gci, check := gci.Find(code, giftCode, user)
			if !check {
				c.DataResponse = iris.Map{"code": -1, "msg": "gift_code_invalid"}
				gc.CheckInput(user,false)
				return
			}
			if !gci.CheckGiftCodeFreedom(user) {
				c.DataResponse = iris.Map{"code": -1, "msg": "gift_code_others_use"}
				gc.CheckInput(user,false)
				return
			}
			gci.UserId = sql.NullString{String: user.UserId, Valid: true}
			gci.ReceiveDate = mysql.NullTime{Time: time.Now(), Valid: true}
			c.DB.Save(&gci)

		case "onlyone":
			gci := models.HfGiftCodeItems{}
			gci, check := gci.Find(code, giftCode, user)
			if !check {
				c.DataResponse = iris.Map{"code": -1, "msg": "gift_code_invalid"}
				gc.CheckInput(user,false)
				return
			}
			if !gci.CheckGiftCodeFreedom(user) {
				c.DataResponse = iris.Map{"code": -1, "msg": "gift_code_others_use"}
				gc.CheckInput(user,false)
				return
			}
			if !gci.CheckGiftCodeOnlyOne(user) {
				c.DataResponse = iris.Map{"code": -1, "msg": "gift_code_i_use"}
				gc.CheckInput(user,false)
				return
			}
			gci.UserId = sql.NullString{String: user.UserId, Valid: true}
			gci.ReceiveDate = mysql.NullTime{Time: time.Now(), Valid: true}
			c.DB.Save(&gci)

		case "all":
			user = c.SetDB(user.ServerId, user)

			ugc := models.HfUserGiftCode{}
			if !ugc.CheckGiftCodeAll(code, user) {
				c.DataResponse = iris.Map{"code": -1, "msg": "gift_code_i_use"}
				gc.CheckInput(user,false)
				return
			}
			ugc.ReceiveDate = time.Now()
			c.DB.Save(&ugc)
		}
		user = c.SetDB(user.ServerId, user)
		gifts := util.JsonDecodeMap(gc.Gift)

		var wg sync.WaitGroup
		wg.Add(1)
		go user.UpdateGifts(gifts, logtype.GIFT_CODE, 0, &wg)

		wg.Add(1)
		dataLog := iris.Map{"server_id": user.ServerId, "user_id": user.UserId, "gift_code": codeInput, "gift": gifts, "created_date": time.Now()}
		go user.SaveLogMongo(logtype.HF_USER_GIFT_CODE_LOG, dataLog, &wg)

		wg.Wait()

		c.DataResponse = iris.Map{"code": 1, "data": iris.Map{constants.GIFT: gifts, "name": gc.Name.String}}
	}
}
