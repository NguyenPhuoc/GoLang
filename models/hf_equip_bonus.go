package models

import (
	"GoLang/libraries/util"
	"encoding/json"
	"fmt"
	"github.com/kataras/iris"
	"sort"
)

type HfEquipBonus struct {
	Quality uint8
	Star    uint8
	Name    string
	Bonus1  string
	Bonus2  string
	Bonus3  string
}

func (HfEquipBonus) TableName() string {
	return "hf_equip_bonus"
}

func (eb *HfEquipBonus) GetMap() iris.Map {
	return iris.Map{
		"quality": eb.Quality,
		"star":    eb.Star,
		"name":    eb.Name,
		"bonus1":  util.JsonDecodeMap(eb.Bonus1),
		"bonus2":  util.JsonDecodeMap(eb.Bonus2),
		"bonus3":  util.JsonDecodeMap(eb.Bonus3),
	}
}

func (eb *HfEquipBonus) GetAll(u HfUser) []HfEquipBonus {
	results := []HfEquipBonus{}

	check := u.RedisConfig.Exists(eb.TableName())
	if check.Val() == 0 {
		results = eb.CacheAll(u)
	} else {
		cacheValue := u.RedisConfig.HGetAll(eb.TableName())
		for _, val := range cacheValue.Val() {
			obj := HfEquipBonus{}

			_ = json.Unmarshal([]byte(val), &obj)
			results = append(results, obj)
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Quality < results[j].Quality || results[i].Quality == results[j].Quality && results[i].Star < results[j].Star
	})

	return results
}

func (eb *HfEquipBonus) CacheAll(u HfUser) []HfEquipBonus {
	results := []HfEquipBonus{}
	u.DB.Find(&results)

	u.RedisConfig.Del(eb.TableName())

	pipe := u.RedisConfig.Pipeline()
	for _, val := range results {
		field := fmt.Sprintf("%d_%d", val.Quality, val.Star)
		pipe.HSet(eb.TableName(), field, util.JsonEndCode(val))
	}
	_, _ = pipe.Exec()
	return results
}
