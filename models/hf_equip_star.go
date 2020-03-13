package models

import (
	"GoLang/libraries/util"
	"encoding/json"
	"fmt"
	"sort"
)

type HfEquipStar struct {
	Quality  uint8 `gorm:"primary_key"`
	Star     uint8 `gorm:"primary_key"`
	Gold     uint
	Quantity uint16
	Sell     uint
}

func (HfEquipStar) TableName() string {
	return "hf_equip_star"
}

func (es *HfEquipStar) Find(quality uint8, star uint8, u HfUser) (HfEquipStar, bool) {
	es.Quality = quality
	es.Star = star
	field := fmt.Sprintf("%d_%d", es.Quality, es.Star)

	count := 0
	if quality != 0 && star != 0 {
		cacheValue := u.RedisConfig.HGet(es.TableName(), field)

		if cacheValue.Err() != nil {
			u.DB.Where(es).First(&es).Count(&count)

			if count != 0 {
				u.RedisConfig.HSet(es.TableName(), field, util.JsonEndCode(es))
			}
		} else {
			count = 1
			_ = json.Unmarshal([]byte(cacheValue.Val()), &es)
		}
	}

	return *es, count != 0
}

func (es *HfEquipStar) GetAll(u HfUser) []HfEquipStar {
	results := []HfEquipStar{}

	check := u.RedisConfig.Exists(es.TableName())
	if check.Val() == 0 {
		results = es.CacheAll(u)
	} else {
		cacheValue := u.RedisConfig.HGetAll(es.TableName())
		for _, val := range cacheValue.Val() {
			obj := HfEquipStar{}

			_ = json.Unmarshal([]byte(val), &obj)
			results = append(results, obj)
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Quality < results[j].Quality || results[i].Quality == results[j].Quality && results[i].Star < results[j].Star
	})

	return results
}

func (es *HfEquipStar) CacheAll(u HfUser) []HfEquipStar {
	results := []HfEquipStar{}
	u.DB.Find(&results)

	u.RedisConfig.Del(es.TableName())

	pipe := u.RedisConfig.Pipeline()
	for _, val := range results {
		field := fmt.Sprintf("%d_%d", val.Quality, val.Star)
		pipe.HSet(es.TableName(), field, util.JsonEndCode(val))
	}
	_, _ = pipe.Exec()
	return results
}
