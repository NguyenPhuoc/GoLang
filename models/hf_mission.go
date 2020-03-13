package models

import (
	"GoLang/libraries/util"
	"encoding/json"
	"fmt"
	"sort"
)

type HfMission struct {
	Key        string `gorm:"primary_key"`
	MissionId  int    `gorm:"primary_key"`
	MissionKey string
	Name       string
	Progress   int
	Gift       string
}

func (HfMission) TableName() string {
	return "hf_mission"
}

func (m *HfMission) GetMission(key string, missionId int, u HfUser) (HfMission, bool) {
	m.Key = key
	m.MissionId = missionId
	field := fmt.Sprintf("%s_%d", key, missionId)

	count := 0
	cacheValue := u.RedisConfig.HGet(m.TableName(), field)

	if cacheValue.Err() != nil {
		results := []HfMission{};
		u.DB.Where(m).First(&results).Count(&count)

		if count != 0 {
			*m = results[0]
			u.RedisConfig.HSet(m.TableName(), field, util.JsonEndCode(m))
		}
	} else {
		count = 1
		_ = json.Unmarshal([]byte(cacheValue.Val()), &m)
	}

	return *m, count != 0
}

func (m *HfMission) GetProgressDaily(u HfUser) ([]HfMission, map[int][]int, map[int]int) {
	allMission := m.GetAll(u)

	progressMission := map[int][]int{}
	progressGift := map[int]int{}
	dailyMission := []HfMission{}

	for _, val := range allMission {
		if val.Key == "daily" {
			progressMission[val.MissionId] = []int{0: 0, 1: val.Progress}
			progressGift[val.MissionId] = 2

			dailyMission = append(dailyMission, val)
		}
	}

	return dailyMission, progressMission, progressGift
}

func (m *HfMission) GetBigGiftDaily(u HfUser) map[string]interface{} {

	m.GetMission("daily_gift", 0, u)

	return util.JsonDecodeMap(m.Gift)
}

func (m *HfMission) GetGiftDaily(missionId int, u HfUser) map[string]interface{} {

	m.GetMission("daily", missionId, u)

	return util.JsonDecodeMap(m.Gift)
}

func (m *HfMission) GetAll(u HfUser) []HfMission {
	results := []HfMission{}

	check := u.RedisConfig.Exists(m.TableName())
	if check.Val() == 0 {
		results = m.CacheAll(u)
	} else {
		cacheValue := u.RedisConfig.HGetAll(m.TableName())
		for _, val := range cacheValue.Val() {
			obj := HfMission{}

			_ = json.Unmarshal([]byte(val), &obj)
			results = append(results, obj)
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Key < results[j].Key || results[i].Key == results[j].Key && results[i].MissionId < results[j].MissionId
	})

	return results
}

func (m *HfMission) CacheAll(u HfUser) []HfMission {
	results := []HfMission{}
	u.DB.Find(&results)

	u.RedisConfig.Del(m.TableName())

	pipe := u.RedisConfig.Pipeline()
	for _, val := range results {
		field := fmt.Sprintf("%s_%d", val.Key, val.MissionId)
		pipe.HSet(m.TableName(), field, util.JsonEndCode(val))
	}
	_, _ = pipe.Exec()
	return results
}
