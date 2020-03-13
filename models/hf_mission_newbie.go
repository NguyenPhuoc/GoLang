package models

import (
	"GoLang/libraries/util"
	"encoding/json"
	"fmt"
	"sort"
)

type HfMissionNewbie struct {
	Key        string `gorm:"primary_key"`
	MissionId  int    `gorm:"primary_key"`
	MissionKey string
	Name       string
	Progress   int
	Gift       string
	Note       string
}

func (HfMissionNewbie) TableName() string {
	return "hf_mission_newbie"
}

func (mn *HfMissionNewbie) GetMission(key string, missionId int, u HfUser) (HfMissionNewbie, bool) {
	mn.Key = key
	mn.MissionId = missionId
	field := fmt.Sprintf("%s_%d", key, missionId)

	count := 0
	cacheValue := u.RedisConfig.HGet(mn.TableName(), field)

	if cacheValue.Err() != nil {
		results := []HfMissionNewbie{};
		u.DB.Where(mn).First(&results).Count(&count)

		if count != 0 {
			*mn = results[0]
			u.RedisConfig.HSet(mn.TableName(), field, util.JsonEndCode(mn))
		}
	} else {
		count = 1
		_ = json.Unmarshal([]byte(cacheValue.Val()), &mn)
	}

	return *mn, count != 0
}

func (mn *HfMissionNewbie) GetProgress(u HfUser) (dailyMission []HfMissionNewbie, progressMission map[int][]int, progressGift map[int]int, progressStep map[int]int) {
	allMisson := mn.GetAll(u)

	dailyMission = []HfMissionNewbie{}
	progressMission = map[int][]int{}
	progressGift = map[int]int{}
	progressStep = map[int]int{}

	for _, val := range allMisson {
		if val.Key == "mission" {
			progressMission[val.MissionId] = []int{0: 0, 1: val.Progress}
			progressGift[val.MissionId] = 2

			dailyMission = append(dailyMission, val)
		} else if val.Key == "step" {

			progressStep[val.MissionId] = 0
		}
	}

	return //dailyMission, progressMission, progressGift, progressStep
}

func (mn *HfMissionNewbie) GetBigGift(u HfUser) map[string]interface{} {

	mn.GetMission("gift", 0, u)

	return util.JsonDecodeMap(mn.Gift)
}

func (mn *HfMissionNewbie) GetMissionGift(missionId int, u HfUser) map[string]interface{} {

	mn.GetMission("mission", missionId, u)

	return util.JsonDecodeMap(mn.Gift)
}

func (mn *HfMissionNewbie) GetStepGift(missionId int, u HfUser) map[string]interface{} {

	mn.GetMission("step", missionId, u)

	return util.JsonDecodeMap(mn.Gift)
}

func (mn *HfMissionNewbie) GetAll(u HfUser) []HfMissionNewbie {
	results := []HfMissionNewbie{}

	check := u.RedisConfig.Exists(mn.TableName())
	if check.Val() == 0 {
		results = mn.CacheAll(u)
	} else {
		cacheValue := u.RedisConfig.HGetAll(mn.TableName())
		for _, val := range cacheValue.Val() {
			obj := HfMissionNewbie{}

			_ = json.Unmarshal([]byte(val), &obj)
			results = append(results, obj)
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Key < results[j].Key || results[i].Key == results[j].Key && results[i].MissionId < results[j].MissionId
	})

	return results
}

func (mn *HfMissionNewbie) CacheAll(u HfUser) []HfMissionNewbie {
	results := []HfMissionNewbie{}
	u.DB.Find(&results)

	u.RedisConfig.Del(mn.TableName())

	pipe := u.RedisConfig.Pipeline()
	for _, val := range results {
		field := fmt.Sprintf("%s_%d", val.Key, val.MissionId)
		pipe.HSet(mn.TableName(), field, util.JsonEndCode(val))
	}
	_, _ = pipe.Exec()
	return results
}
