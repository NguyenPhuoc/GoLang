package models

import (
	"GoLang/libraries/constants"
	"GoLang/libraries/util"
	"database/sql"
	"encoding/json"
	"github.com/go-sql-driver/mysql"
	"github.com/jinzhu/now"
	"github.com/kataras/iris"
	"sort"
	"time"
)

type HfEvent struct {
	Id         int `gorm:"primary_key"`
	Name       string
	Type       string
	StartDate  mysql.NullTime
	EndDate    mysql.NullTime
	Note       sql.NullString
	DeviceHide string
	Config     string
}

func (HfEvent) TableName() string {
	return "hf_event"
}

func (e HfEvent) GetMap(u HfUser) iris.Map {
	isNotify := 0

	ue := HfUserEvent{}
	switch e.Type {
	case "newbie_rollup":
		ue = ue.GetNewbieRollup(u)
		progress := map[int]int{}
		util.JsonDecodeObject(ue.Progress.String, &progress)
		for _, val := range progress {
			if val == 0 {
				isNotify = 1
				break
			}
		}
	case "p8":
		ue, check := ue.Get(8, u)
		//có, chưa nhận ngày, chưa hết hạn
		if check && now.New(ue.UpdateDate.Time).EndOfDay() != now.EndOfDay() && ue.ReceiveDate.Time.Unix() > time.Now().Unix() {
			isNotify = 1
		}
	case "p9":
		ue, check := ue.Get(9, u)
		//có, chưa nhận ngày, chưa hết hạn
		if check && now.New(ue.UpdateDate.Time).EndOfDay() != now.EndOfDay() && ue.ReceiveDate.Time.Unix() > time.Now().Unix() {
			isNotify = 1
		}
	case "p10":
		ue = ue.GetGrowUp(u)
		progress := map[int]int{}
		progressBonus := map[int]int{}
		util.JsonDecodeObject(ue.Progress.String, &progress)
		util.JsonDecodeObject(ue.Progress.String, &progressBonus)
		for i, _ := range progress {
			if progress[i] == 0 || progressBonus[i] == 0 {
				isNotify = 1
				break
			}
		}
	case "payment_everyday":
		ue = ue.GetPaymentEveryday(u)
		progress := map[uint]int{}
		util.JsonDecodeObject(ue.Progress.String, &progress)
		for _, val := range progress {
			if val == 0 {
				isNotify = 1
				break
			}
		}
	case "payment_accumulate":
		ue = ue.GetPaymentAccumulate(u)
		progress := map[uint]int{}
		util.JsonDecodeObject(ue.Progress.String, &progress)
		for _, val := range progress {
			if val == 0 {
				isNotify = 1
				break
			}
		}
	case "turn_basic":
		ue = ue.GetTurnBasic(u)
		progress := map[uint]int{}
		util.JsonDecodeObject(ue.Progress.String, &progress)
		for _, val := range progress {
			if val == 0 {
				isNotify = 1
				break
			}
		}
	case "turn_basic_v2":
		ue, _, check := ue.GetTurnBasicV2(u)
		if check {
			progress := map[uint]int{}
			util.JsonDecodeObject(ue.Progress.String, &progress)
			for _, val := range progress {
				if val == 0 {
					isNotify = 1
					break
				}
			}
		}
	}

	return iris.Map{
		"id":         e.Id,
		"name":       e.Name,
		"type":       e.Type,
		"start_date": e.StartDate.Time.Unix(),
		"end_date":   e.EndDate.Time.Unix(),
		"note":       e.Note.String,
		"is_notify":  isNotify,
		"config":     util.JsonDecodeMap(e.Config),
	}
}

func (e *HfEvent) Find(eventId int, u HfUser) (HfEvent, bool) {
	e.Id = eventId
	field := util.ToString(eventId)

	count := 0
	if eventId != 0 {
		cacheValue := u.RedisConfig.HGet(e.TableName(), field)
		if cacheValue.Err() != nil {
			u.DB.First(&e).Count(&count)
			e.CacheAll(u)

			if count != 0 {
				u.RedisConfig.HSet(e.TableName(), field, util.JsonEndCode(e))
			}
		} else {
			count = 1
			_ = json.Unmarshal([]byte(cacheValue.Val()), &e)
		}
	}

	if count == 1 {
		if (e.StartDate.Valid || e.EndDate.Valid) && (time.Now().Unix() < e.StartDate.Time.Unix() || time.Now().Unix() > e.EndDate.Time.Unix()) {
			count = 0
		}
	}

	return *e, count != 0
}

func (e *HfEvent) FindType(eventType string, u HfUser) (HfEvent, bool) {

	events := e.GetAll(u)
	count := 0
	for _, event := range events {
		if event.Type == eventType &&
			((e.StartDate.Valid && e.EndDate.Valid && time.Now().Unix() >= event.StartDate.Time.Unix() && time.Now().Unix() <= event.EndDate.Time.Unix()) ||
				(!e.StartDate.Valid || !e.EndDate.Valid)) {

			count = 1
			*e = event
			break
		}
	}

	return *e, count != 0
}

func (e *HfEvent) GetAll(u HfUser) []HfEvent {
	results := []HfEvent{}

	check := u.RedisConfig.Exists(e.TableName())
	if check.Val() == 0 {
		results = e.CacheAll(u)
	} else {
		cacheValue := u.RedisConfig.HGetAll(e.TableName())
		for _, val := range cacheValue.Val() {
			obj := HfEvent{}

			_ = json.Unmarshal([]byte(val), &obj)
			results = append(results, obj)
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Id < results[j].Id
	})

	return results
}

func (e *HfEvent) CacheAll(u HfUser) []HfEvent {
	results := []HfEvent{}
	u.DB.Find(&results)

	u.RedisConfig.Del(e.TableName())

	pipe := u.RedisConfig.Pipeline()
	for _, val := range results {
		pipe.HSet(e.TableName(), util.ToString(val.Id), util.JsonEndCode(val))
	}
	_, _ = pipe.Exec()
	return results
}

func (e *HfEvent) GetProgressInt() map[int]int {
	gift_cf := map[int]iris.Map{}
	util.JsonDecodeObject(e.Config, &gift_cf)

	progressReturn := map[int]int{}
	progress := util.MapKeysInt(gift_cf, constants.ASC)
	for _, v := range progress {
		progressReturn[v] = 2
	}

	return progressReturn
}
