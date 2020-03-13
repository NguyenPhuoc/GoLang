package models

import (
	"GoLang/libraries/util"
	"encoding/json"
	"sort"
)

type HfArenaShop struct {
	Id        int `gorm:"primary_key"`
	Name      string
	Price     uint
	TypePrice string
	Cost      uint
	Limit     int8
	Gift      string
}

func (HfArenaShop) TableName() string {
	return "hf_arena_shop"
}

func (as *HfArenaShop) Find(id int, u HfUser) (HfArenaShop, bool) {
	as.Id = id

	cacheValue := u.RedisConfig.HGet(as.TableName(), util.ToString(id))

	count := 0
	if cacheValue.Err() != nil {
		u.DB.Where(as).First(&as).Count(&count)

		if count != 0 {
			u.RedisConfig.HSet(as.TableName(), util.ToString(id), util.JsonEndCode(as))
		}
	} else {
		count = 1
		_ = json.Unmarshal([]byte(cacheValue.Val()), &as)
	}

	return *as, count != 0
}

func (as *HfArenaShop) GetAll(u HfUser) []HfArenaShop {
	results := []HfArenaShop{}

	check := u.RedisConfig.Exists(as.TableName())
	if check.Val() == 0 {
		results = as.CacheAll(u)
	} else {
		cacheValue := u.RedisConfig.HGetAll(as.TableName())
		for _, val := range cacheValue.Val() {
			obj := HfArenaShop{}

			_ = json.Unmarshal([]byte(val), &obj)
			results = append(results, obj)
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Id < results[j].Id
	})

	return results
}

func (as *HfArenaShop) CacheAll(u HfUser) []HfArenaShop {
	results := []HfArenaShop{}
	u.DB.Find(&results)

	u.RedisConfig.Del(as.TableName())

	pipe := u.RedisConfig.Pipeline()
	for _, val := range results {
		pipe.HSet(as.TableName(), util.ToString(val.Id), util.JsonEndCode(val))
	}
	_, _ = pipe.Exec()
	return results
}
