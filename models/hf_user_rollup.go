package models

import (
	"GoLang/libraries/constants"
	"GoLang/libraries/util"
	"encoding/json"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"github.com/jinzhu/now"
	"github.com/kataras/iris"
	"sync"
	"time"
)

type HfUserRollup struct {
	Month       time.Time `gorm:"primary_key"`
	UserId      string    `gorm:"primary_key"`
	Total       uint8
	Progress    string
	TimeReceive mysql.NullTime
}

func (HfUserRollup) TableName() string {
	return "hf_user_rollup"
}

func (ur *HfUserRollup) GetMap(u HfUser) iris.Map {
	isReceive := 0
	if ur.TimeReceive.Valid == true && ur.TimeReceive.Time.Day() == time.Now().Day() {
		isReceive = 1
	}

	rollup := HfRollup{}
	rollup.Get(ur.Month, u)
	giftCf := util.JsonDecodeMap(rollup.Gift)

	progress := util.JsonDecodeMap(ur.Progress)
	progressString := map[string]iris.Map{}
	for day, status := range progress {

		progressString[day] = iris.Map{
			constants.GIFT: giftCf[day],
			"status":       util.StatusGift(status),
		}
	}

	nextMonth := now.EndOfMonth().Unix() - time.Now().Unix()

	return iris.Map{
		"total":      ur.Total,
		"progress":   progressString,
		"status":     util.StatusGift(isReceive),
		"next_month": nextMonth,
	}
}

func (ur *HfUserRollup) UpdateCache(u HfUser) {
	field := fmt.Sprintf("%s_%d_%02d", ur.TableName(), ur.Month.Year(), ur.Month.Month())
	u.RedisInfo.HSet(ur.UserId, field, util.JsonEndCode(ur))
}

func (ur *HfUserRollup) Save(u HfUser, wg *sync.WaitGroup) {
	defer wg.Done()
	ur.UpdateCache(u)
	u.DB.Save(&ur)
}

func (ur *HfUserRollup) Get(u HfUser) HfUserRollup {
	month := now.BeginningOfMonth()
	ur.Month = month
	ur.UserId = u.UserId
	field := fmt.Sprintf("%s_%d_%02d", ur.TableName(), month.Year(), month.Month())

	cacheValue := u.RedisInfo.HGet(u.UserId, field)

	if cacheValue.Err() != nil {
		count := 0
		results := []HfUserRollup{};
		u.DB.Where(ur).First(&results).Count(&count)

		if count != 0 {
			*ur = results[0]
			u.RedisInfo.HSet(u.UserId, field, util.JsonEndCode(ur))
		} else {
			lastDay := now.EndOfMonth().Day()
			progress := map[int]int{}
			for i := 1; i <= lastDay; i++ {
				progress[i] = 2
			}

			ur.Progress = util.JsonEndCode(progress)
			var wg sync.WaitGroup
			wg.Add(1)
			go ur.Save(u, &wg)
			wg.Wait()
		}
	} else {
		_ = json.Unmarshal([]byte(cacheValue.Val()), &ur)
	}

	ur.CheckDay(u)

	return *ur
}

func (ur *HfUserRollup) CheckDay(u HfUser) {

	if ur.TimeReceive.Valid == false || ur.TimeReceive.Time.Day() != time.Now().Day() {

		day := util.ToString(ur.Total + 1)
		progress := util.JsonDecodeMap(ur.Progress)
		if val, ok := progress[day]; ok && util.ToInt(val) == 2 {

			progress[day] = 0
			ur.Progress = util.JsonEndCode(progress)
		}
	}
}
