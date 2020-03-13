package models

import (
	"GoLang/libraries/constants"
	"GoLang/libraries/logtype"
	"GoLang/libraries/util"
	"database/sql"
	"github.com/jinzhu/gorm"
	"math"
	"sync"
	"time"
)

type HfUserTower struct {
	UserId      string `gorm:"primary_key"`
	FullName    string `gorm:"default:NULL"`
	Floor       int    `gorm:"default:1"`
	Power       uint   `gorm:"default:10"`
	CheckPower  int
	PowerPoint  int
	Level       int
	UpdatePower time.Time
	LastLineUp  sql.NullString
}

func (HfUserTower) TableName() string {
	return "hf_user_tower"
}

func (HfUserTower) MaxPower() uint {
	return 10
}

func (HfUserTower) BlockTime() int {
	return 3600 * 2
}
func (HfUserTower) GetPricePower(quantity int) int {
	//Giá của mỗi power là 10 gem
	return quantity * 10
}

func (ut *HfUserTower) Get(u HfUser) HfUserTower {
	ut.UserId = u.UserId

	count := 0
	u.DB.Where(ut).First(&ut).Count(&count)
	if count == 0 {
		ut.FullName = u.FullName
		ut.Floor = 1
		ut.Power = 10
		ut.UpdatePower = time.Now()
		u.DB.Save(&ut)
	}

	return *ut
}

func (ut *HfUserTower) IncreasePower(u HfUser) HfUserTower {
	if ut.Power >= ut.MaxPower() {
		return *ut
	}
	times := int(math.Max(float64(time.Now().Unix()-ut.UpdatePower.Unix()), 0))
	power := int(times / ut.BlockTime()) // 2h/power

	if power == 0 {
		return *ut
	}

	powerWillBe := uint(power) + ut.Power
	if powerWillBe >= ut.MaxPower() {
		power = int(ut.MaxPower() - ut.Power)
	}
	ut.UpdatePower = time.Now()

	var wg sync.WaitGroup
	wg.Add(1)
	go ut.SetPower(power, "", 0, logtype.INCREASE_POWER_TOWER, 0, u, &wg)
	wg.Wait()

	return *ut
}

func (ut *HfUserTower) SetPower(quantity int, itemId interface{}, kindId interface{}, typeLog, eventId int, u HfUser, wg *sync.WaitGroup) {
	defer wg.Done()

	ut.Power = util.QuantityUint(ut.Power, quantity)

	wg.Add(2)
	go u.SaveLog(typeLog, constants.POWER_TOWER, itemId, kindId, eventId, quantity, uint64(ut.Power), "", wg)

	go func() {
		u.DB.Save(&ut)
		wg.Done()
	}()
}

func (ut *HfUserTower) GetRankAndTop(db *gorm.DB) (int, string) {

	tops := []struct {
		UserId     string `json:"user_id"`
		FullName   string `json:"full_name"`
		Floor      int    `json:"floor"`
		PowerPoint int    `json:"power_point"`
		Level      int    `json:"level"`
		Top        int    `json:"top"`
		AvatarId   int    `json:"avatar_id"`
	}{}
	//db.Raw(`SELECT user_id, full_name, floor, power_point, level FROM hf_user_tower ORDER BY floor DESC LIMIT 100;`).Scan(&tops)
	//db.Raw(`SELECT tow.user_id, full_name, floor, power_point, level, ava.avatar_id FROM hf_user_tower tow
	//			LEFT JOIN (SELECT user_id, MAX(avatar_id) avatar_id FROM hf_user_avatar WHERE used = 1 GROUP BY user_id) ava
	//			ON ava.user_id = tow.user_id
	//			WHERE power_point != 0
	//			ORDER BY floor DESC LIMIT 100;`).Scan(&tops)
	db.Raw(`SELECT tow.user_id, tow.full_name, floor, power_point, level, ava.avatar_id FROM hf_user_tower tow 
				LEFT JOIN hf_user ava 
				ON ava.user_id = tow.user_id 
				WHERE power_point != 0 
				ORDER BY floor DESC LIMIT 100;`).Scan(&tops)

	myRank := 0;
	for i, val := range tops {
		val.Top = i + 1
		tops[i].Top = val.Top

		if val.UserId == ut.UserId {
			myRank = val.Top
		}
	}
	if myRank == 0 {
		db.Model(&HfUserTower{}).Where("floor > ?", ut.Floor).Count(&myRank)
		myRank++;
	}

	return myRank, util.JsonEndCode(tops)
}
