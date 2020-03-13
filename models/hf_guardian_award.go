package models

import (
	"GoLang/libraries/util"
	"encoding/json"
	"sort"
)

type HfGuardianAward struct {
	Level    uint16 `gorm:"primary_key"`
	GiftHunt string
	GiftKill string
	Piece    int
}

func (HfGuardianAward) TableName() string {
	return "hf_guardian_award"
}

func (ga *HfGuardianAward) Find(level uint16, u HfUser) (HfGuardianAward, bool) {
	ga.Level = level

	cacheValue := u.RedisConfig.HGet(ga.TableName(), util.ToString(level))

	count := 0
	if cacheValue.Err() != nil {
		u.DB.Where(ga).First(&ga).Count(&count)

		if count != 0 {
			u.RedisConfig.HSet(ga.TableName(), util.ToString(level), util.JsonEndCode(ga))
		}
	} else {
		count = 1
		_ = json.Unmarshal([]byte(cacheValue.Val()), &ga)
	}

	return *ga, count != 0
}

func (ga *HfGuardianAward) GetAll(u HfUser) []HfGuardianAward {
	results := []HfGuardianAward{}

	check := u.RedisConfig.Exists(ga.TableName())
	if check.Val() == 0 {
		results = ga.CacheAll(u)
	} else {
		cacheValue := u.RedisConfig.HGetAll(ga.TableName())
		for _, val := range cacheValue.Val() {
			obj := HfGuardianAward{}

			_ = json.Unmarshal([]byte(val), &obj)
			results = append(results, obj)
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Level < results[j].Level
	})

	return results
}

func (ga *HfGuardianAward) CacheAll(u HfUser) []HfGuardianAward {
	results := []HfGuardianAward{}
	u.DB.Find(&results)

	u.RedisConfig.Del(ga.TableName())

	pipe := u.RedisConfig.Pipeline()
	for _, val := range results {
		pipe.HSet(ga.TableName(), util.ToString(val.Level), util.JsonEndCode(val))
	}
	_, _ = pipe.Exec()
	return results
}
