package models

import (
	"GoLang/libraries/util"
	"encoding/json"
	"sort"
)

type HfPetEvolve struct {
	Evolve uint16 `gorm:"primary_key"`
	Level  uint16
	Stone  uint
	Gold   uint
}

func (HfPetEvolve) TableName() string {
	return "hf_pet_evolve"
}

func (pe *HfPetEvolve) Find(evolve uint16, u HfUser) (HfPetEvolve, bool) {
	pe.Evolve = evolve
	field := util.ToString(evolve)

	count := 0
	if evolve != 0 {
		cacheValue := u.RedisConfig.HGet(pe.TableName(), field)
		if cacheValue.Err() != nil {
			u.DB.First(&pe).Count(&count)

			if count != 0 {
				u.RedisConfig.HSet(pe.TableName(), field, util.JsonEndCode(pe))
			}
		} else {
			count = 1
			_ = json.Unmarshal([]byte(cacheValue.Val()), &pe)
		}
	}

	return *pe, count != 0
}

func (pe *HfPetEvolve) GetAll(u HfUser) []HfPetEvolve {
	results := []HfPetEvolve{}

	check := u.RedisConfig.Exists(pe.TableName())
	if check.Val() == 0 {
		results = pe.CacheAll(u)
	} else {
		cacheValue := u.RedisConfig.HGetAll(pe.TableName())
		for _, val := range cacheValue.Val() {
			obj := HfPetEvolve{}

			_ = json.Unmarshal([]byte(val), &obj)
			results = append(results, obj)
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Evolve < results[j].Evolve
	})

	return results
}

func (pe *HfPetEvolve) CacheAll(u HfUser) []HfPetEvolve {
	results := []HfPetEvolve{}
	u.DB.Find(&results)

	u.RedisConfig.Del(pe.TableName())

	pipe := u.RedisConfig.Pipeline()
	for _, val := range results {
		pipe.HSet(pe.TableName(), util.ToString(val.Evolve), util.JsonEndCode(val))
	}
	_, _ = pipe.Exec()
	return results
}
