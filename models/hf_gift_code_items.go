package models

import (
	"database/sql"
	"github.com/go-sql-driver/mysql"
)

type HfGiftCodeItems struct {
	Code        string `gorm:"primary_key"`
	GiftCode    string `gorm:"primary_key"`
	UserId      sql.NullString
	ReceiveDate mysql.NullTime
}

func (HfGiftCodeItems) TableName() string {
	return "hf_gift_code_items"
}

func (gci *HfGiftCodeItems) Find(code, giftCode string, u HfUser) (HfGiftCodeItems, bool) {
	gci.Code = code
	gci.GiftCode = giftCode

	count := 0
	u.DB.First(&gci).Count(&count)

	return *gci, count != 0
}

func (gci *HfGiftCodeItems) CheckGiftCodeFreedom(u HfUser) (canUse bool) {

	canUse = gci.UserId.Valid == false
	return
}

func (gci *HfGiftCodeItems) CheckGiftCodeOnlyOne(u HfUser) (canUse bool) {

	count := 0
	u.DB.Model(&HfGiftCodeItems{}).Where("code = ? and user_id = ?", gci.Code, u.UserId).Count(&count)

	canUse = count == 0
	return
}
