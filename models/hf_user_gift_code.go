package models

import "time"

type HfUserGiftCode struct {
	UserId      string `gorm:"primary_key"`
	GiftCode    string `gorm:"primary_key"`
	ReceiveDate time.Time
}

func (HfUserGiftCode) TableName() string {
	return "hf_user_gift_code"
}

func (ugc *HfUserGiftCode) CheckGiftCodeAll(giftCode string, u HfUser) (canUse bool) {
	ugc.UserId = u.UserId
	ugc.GiftCode = giftCode

	count := 0
	u.DB.First(&ugc).Count(&count)

	canUse = count == 0
	return
}
