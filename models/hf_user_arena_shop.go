package models

import (
	"GoLang/libraries/constants"
	"GoLang/libraries/logtype"
	"GoLang/libraries/util"
	"github.com/jinzhu/now"
	"github.com/kataras/iris"
	"sync"
	"time"
)

type HfUserArenaShop struct {
	UserId      string `gorm:"primary_key"`
	ArenaCoin   uint
	ItemsBought string
	ResetDate   time.Time
}

func (HfUserArenaShop) TableName() string {
	return "hf_user_arena_shop"
}

func (uas *HfUserArenaShop) Get(u HfUser) HfUserArenaShop {
	uas.UserId = u.UserId

	count := 0
	u.DB.First(&uas).Count(&count)

	if count == 0 {
		uas.ArenaCoin = 0
		uas.ItemsBought = "{}"
		uas.ResetDate = time.Now()
		u.DB.Create(&uas)
	} else if uas.ResetDate.Unix() < now.BeginningOfMonth().Unix() {
		uas.ItemsBought = "{}"
		uas.ResetDate = time.Now()
		u.DB.Save(&uas)
	}

	return *uas
}

func (uas *HfUserArenaShop) SetArenaCoin(arenaCoin int, itemId interface{}, kindId interface{}, typeLog, eventId int, u HfUser, wg *sync.WaitGroup) {
	defer wg.Done()

	uas.ArenaCoin = util.QuantityUint(uas.ArenaCoin, arenaCoin)

	wg.Add(2)
	go u.SaveLog(typeLog, constants.ARENA_COIN, itemId, kindId, eventId, arenaCoin, uint64(uas.ArenaCoin), "", wg)

	go func() {
		u.DB.Save(&uas)
		wg.Done()
	}()
}

func (uas HfUserArenaShop) SaveLog(u HfUser, itemId int, Data interface{}, group *sync.WaitGroup) {
	defer group.Done()

	group.Add(1)
	go u.SaveLogMongo(logtype.HF_USER_ARENA_SHOP_LOG, iris.Map{"server_id": u.ServerId, "user_id": u.UserId, "item_id": itemId, "data": Data, "created_date": time.Now()}, group)
}

func (uas HfUserArenaShop) CostResetConfig(u HfUser) int {
	return util.ToInt(u.GetConfig("arena_shop_cost_reset").Value)
}
