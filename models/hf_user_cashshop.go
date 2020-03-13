package models

import (
	"GoLang/config/cashshop"
	"GoLang/libraries/util"
)

type HfUserCashshop struct {
	UserId       string `gorm:"primary_key"`
	FirstPackage string
	FirstPayment uint8
}

func (HfUserCashshop) TableName() string {
	return "hf_user_cashshop"
}

func (uc *HfUserCashshop) Get(u HfUser) HfUserCashshop {
	uc.UserId = u.UserId

	count := 0
	u.DB.First(&uc).Count(&count)
	if count == 0 {
		firstPackage := map[string]int{}
		cashConfig := cashshop.Config()
		for id, _ := range cashConfig {
			firstPackage[id] = 0
		}
		uc.FirstPackage = util.JsonEndCode(firstPackage)
	}

	return *uc
}
