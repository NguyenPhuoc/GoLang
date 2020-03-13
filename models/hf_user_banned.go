package models

import (
	"sync"
	"time"
)

type HfUserBanned struct {
	UserId     string    `gorm:"primary_key" json:"user_id"`
	FreedomDay time.Time `json:"freedom_day"`
	Reason     string    `json:"reason"`
}

func (HfUserBanned) TableName() string {
	return "hf_user_banned"
}

func (ub *HfUserBanned) Find(u HfUser) (HfUserBanned, bool) {
	ub.UserId = u.UserId

	count := 0
	u.DB.First(&ub).Count(&count)

	return *ub, count != 0
}

func (ub *HfUserBanned) UnBan(u HfUser){
	ub.UserId = u.UserId

	u.DB.Delete(&ub)

	var wg sync.WaitGroup
	wg.Add(1)
	u.Banned = 0
	go u.UpdateKey("Banned", u.Banned, &wg)
	wg.Wait()
}
