package models

import (
	"GoLang/libraries/util"
	"encoding/json"
	"fmt"
	"sort"
)

type HfGuardianUpgrade struct {
	GuardianId           uint16 `gorm:"primary_key"`
	Evolve               uint16 `gorm:"primary_key"`
	Piece                uint
	FruitFormula         string
	GoldFormula          string
	FlowerFormulaPassive string
	GoldFormulaPassive   string
}

func (HfGuardianUpgrade) TableName() string {
	return "hf_guardian_upgrade"
}

func (gu *HfGuardianUpgrade) Find(guardianId, evolve uint16, u HfUser) (HfGuardianUpgrade, bool) {
	field := fmt.Sprintf("%d_%d", guardianId, evolve)

	count := 0
	cacheValue := u.RedisConfig.HGet(gu.TableName(), field)

	if cacheValue.Err() != nil {
		u.DB.Where("guardian_id = ? AND evolve = ?", guardianId, evolve).First(&gu).Count(&count)

		if count != 0 {
			u.RedisConfig.HSet(gu.TableName(), field, util.JsonEndCode(gu))
		}
	} else {
		count = 1
		_ = json.Unmarshal([]byte(cacheValue.Val()), &gu)
	}

	return *gu, count != 0
}

func (gu *HfGuardianUpgrade) GetAll(u HfUser) []HfGuardianUpgrade {
	results := []HfGuardianUpgrade{}

	check := u.RedisConfig.Exists(gu.TableName())
	if check.Val() == 0 {
		results = gu.CacheAll(u)
	} else {
		cacheValue := u.RedisConfig.HGetAll(gu.TableName())
		for _, val := range cacheValue.Val() {
			obj := HfGuardianUpgrade{}

			_ = json.Unmarshal([]byte(val), &obj)
			results = append(results, obj)
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].GuardianId < results[j].GuardianId || results[i].GuardianId == results[j].GuardianId && results[i].Evolve < results[j].Evolve
	})

	return results
}

func (gu *HfGuardianUpgrade) CacheAll(u HfUser) []HfGuardianUpgrade {
	results := []HfGuardianUpgrade{}
	u.DB.Find(&results)

	u.RedisConfig.Del(gu.TableName())

	pipe := u.RedisConfig.Pipeline()
	for _, val := range results {
		field := fmt.Sprintf("%d_%d", val.GuardianId, val.Evolve)
		pipe.HSet(gu.TableName(), field, util.JsonEndCode(val))
	}
	_, _ = pipe.Exec()
	return results
}
