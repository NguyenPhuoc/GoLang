package models

import (
	"GoLang/libraries/util"
	"encoding/json"
	"sort"
)

type HfPetLevel struct {
	Level  uint16 `gorm:"primary_key"`
	Evolve uint16
	Stone  uint
	Gold   uint
}

func (HfPetLevel) TableName() string {
	return "hf_pet_level"
}

func (pl *HfPetLevel) GetLevel(level uint16, u HfUser) (HfPetLevel, bool) {
	pl.Level = level

	field := util.ToString(level)

	count := 0
	if level != 0 {
		cacheValue := u.RedisConfig.HGet(pl.TableName(), field)
		if cacheValue.Err() != nil {
			u.DB.Where(pl).First(&pl).Count(&count)

			if count != 0 {
				u.RedisConfig.HSet(pl.TableName(), field, util.JsonEndCode(pl))
			}
		} else {
			count = 1
			_ = json.Unmarshal([]byte(cacheValue.Val()), &pl)
		}
	}

	return *pl, count != 0
}

func (pl *HfPetLevel) GetMaxLevel() map[uint8]uint16 {
	return map[uint8]uint16{1: 40, 2: 60, 3: 80, 4: 100, 5: 100};
}

func (pl *HfPetLevel) GetAll(u HfUser) []HfPetLevel {
	results := []HfPetLevel{}

	check := u.RedisConfig.Exists(pl.TableName())
	if check.Val() == 0 {
		results = pl.CacheAll(u)
	} else {
		cacheValue := u.RedisConfig.HGetAll(pl.TableName())
		for _, val := range cacheValue.Val() {
			obj := HfPetLevel{}

			_ = json.Unmarshal([]byte(val), &obj)
			results = append(results, obj)
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Level < results[j].Level
	})

	return results
}

func (pl *HfPetLevel) CacheAll(u HfUser) []HfPetLevel {
	results := []HfPetLevel{}
	u.DB.Find(&results)

	u.RedisConfig.Del(pl.TableName())

	pipe := u.RedisConfig.Pipeline()
	for _, val := range results {
		pipe.HSet(pl.TableName(), util.ToString(val.Level), util.JsonEndCode(val))
	}
	_, _ = pipe.Exec()
	return results
}
