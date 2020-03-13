package controllers

import (
	"GoLang/libraries/constants"
	"GoLang/libraries/util"
	"GoLang/models"
	"database/sql"
	"github.com/go-sql-driver/mysql"
	"github.com/kataras/iris"
	"sync"
	"time"
)

type InboxController struct {
	MyController
}

// /inbox/list
func (c *InboxController) GetList(ctx iris.Context) {
	user := c.User

	results := []models.HfUserInbox{}
	total := 0

	uInbox := models.HfUserInbox{ReceiverId: user.UserId}

	c.DB.Where(&uInbox).Where("is_delete = 0").Order("created_date desc").Find(&results).Count(&total)

	resultsMap := make([]iris.Map, len(results))
	for i, val := range results {
		resultsMap[i] = val.GetMap()
	}

	c.DataResponse = iris.Map{"code": 1, "data": iris.Map{"inbox": resultsMap, "total": total}}
}

// /inbox/read
func (c *InboxController) PostRead(form formValue, ctx iris.Context) {
	if c.validToken(ctx) {
		user := c.User
		id := form("id")

		check := 0
		uInbox := models.HfUserInbox{ReceiverId: user.UserId, Id: id}
		c.DB.Where(&uInbox).Where("is_read = 0").First(&uInbox).Count(&check)

		if check == 0 {
			c.DataResponse = iris.Map{"code": -1, "msg": "Invalid"}
			return
		}

		uInbox.IsRead = 1
		uInbox.ReadDate = mysql.NullTime{Time: time.Now(), Valid: true}
		c.DB.Save(&uInbox)

		c.DataResponse = iris.Map{"code": 1, "msg": "Success"}
	}
}

// /inbox/receive
func (c *InboxController) PostReceive(form formValue, ctx iris.Context) {
	if c.validToken(ctx) {
		user := c.User
		id := form("id")

		check := 0
		uInbox := models.HfUserInbox{ReceiverId: user.UserId, Id: id}
		c.DB.Where(&uInbox).Where("is_read = 1 and is_receive = 0").First(&uInbox).Count(&check)

		if check == 0 {
			c.DataResponse = iris.Map{"code": -1, "msg": "Invalid"}
			return
		}

		var wg sync.WaitGroup
		wg.Add(2)

		gifts := util.JsonDecodeMap(uInbox.Gift.String)

		go user.UpdateGifts(gifts, uInbox.TypeLog, uInbox.EventId, &wg)

		uInbox.IsReceive = 1
		uInbox.ReceivedDate = mysql.NullTime{Time: time.Now(), Valid: true}
		go func() {
			c.DB.Save(&uInbox)
			wg.Done()
		}()
		wg.Wait()

		c.DataResponse = iris.Map{"code": 1, "msg": "Success", "data": iris.Map{constants.GIFT: gifts}}
	}
}

// /inbox/receive/all
func (c *InboxController) PostReceiveAll(ctx iris.Context) {
	if c.validToken(ctx) {
		user := c.User

		results := []models.HfUserInbox{}
		uInbox := models.HfUserInbox{ReceiverId: user.UserId}
		c.DB.Where(&uInbox).Where("is_receive = 0").Find(&results)

		if len(results) == 0 {
			c.DataResponse = iris.Map{"code": -1, "msg": "Invalid"}
			return
		}

		var wg sync.WaitGroup
		gifts := map[string]interface{}{}
		for _, uIn := range results {
			wg.Add(2)
			gift := util.JsonDecodeMap(uIn.Gift.String)
			gifts = util.MergeGift(gifts, gift)

			go user.UpdateGifts(gift, uIn.TypeLog, uIn.EventId, &wg)
			go func() {
				uIn.IsRead = 1
				uIn.IsReceive = 1
				uIn.ReadDate = mysql.NullTime{Time: time.Now(), Valid: true}
				uIn.ReceivedDate = mysql.NullTime{Time: time.Now(), Valid: true}
				c.DB.Save(&uIn)
				wg.Done()
			}()
			wg.Wait()
		}

		c.DataResponse = iris.Map{"code": 1, "msg": "Success", "data": iris.Map{constants.GIFT: gifts}}
	}
}

// /inbox/delete
func (c *InboxController) PostDelete(form formValue, ctx iris.Context) {
	if c.validToken(ctx) {
		user := c.User
		id := form("id")

		check := 0
		uInbox := models.HfUserInbox{ReceiverId: user.UserId, Id: id}
		c.DB.Where(&uInbox).Where("is_read = 1 and is_receive != 0 and is_delete = 0").First(&uInbox).Count(&check)

		if check == 0 {
			c.DataResponse = iris.Map{"code": -1, "msg": "Invalid"}
			return
		}

		uInbox.IsDelete = 1
		uInbox.DeletedDate = mysql.NullTime{Time: time.Now(), Valid: true}
		c.DB.Save(&uInbox)

		c.DataResponse = iris.Map{"code": 1, "msg": "Success"}
	}
}

// /inbox/delete/all
func (c *InboxController) PostDeleteAll(ctx iris.Context) {
	if c.validToken(ctx) {
		user := c.User

		uInbox := models.HfUserInbox{}
		c.DB.Model(&uInbox).Where("receiver_id = ? and is_read = 1 and is_receive != 0 and is_delete = 0", user.UserId).Updates(map[string]interface{}{"is_delete": 1, "deleted_date": mysql.NullTime{Time: time.Now(), Valid: true}})

		c.DataResponse = iris.Map{"code": 1, "msg": "Success"}
	}
}

// /inbox/send/message
func (c *InboxController) PostSendMessage(form formValue, ctx iris.Context) {
	if c.validToken(ctx) {
		user := c.User

		userId := form("user_id")
		serverId := uint(util.ToInt(form("server_id")))
		message := form("message")

		c.SetDB(serverId)
		u := models.HfUser{}
		count := 0
		c.DB.Where("user_id = ? and server_id = ?", userId, serverId).First(&u).Count(&count)
		if count == 0 {
			c.DataResponse = iris.Map{"code": -1, "msg": "Can't found user"}
			return
		}

		title := message
		titleLen := 10
		if len([]rune(message)) > titleLen {
			title = string([]rune(message)[:titleLen])
			title += "..."
		}

		uInbox := models.HfUserInbox{
			Id:          util.UUID(),
			ReceiverId:  userId,
			SenderId:    sql.NullString{String: user.UserId, Valid: true},
			FullName:    sql.NullString{String: user.FullName, Valid: true},
			ServerId:    user.ServerId,
			SenderType:  constants.INBOX_SENDER_BY_MESSAGE,
			IsReceive:   2,
			Title:       "Lời Nhắn Bạn Bè",
			Message:     sql.NullString{String: message, Valid: true},
			CreatedDate: time.Now(),
		}
		c.DB.Save(&uInbox)

		c.DataResponse = iris.Map{"code": 1, "msg": "Success"}
	}
}
