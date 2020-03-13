package models

import (
	"GoLang/libraries/util"
	"encoding/json"
	"sort"
)

type HfWhitelist struct {
	Id       int    `gorm:"primary_key" json:"id"`
	Value    string `json:"value"`
	Type     string `json:"type"`
	ServerId uint   `json:"server_id"`
}

func (HfWhitelist) TableName() string {
	return "hf_whitelist"
}

func (wl *HfWhitelist) GetAll(u HfUser) []HfWhitelist {
	results := []HfWhitelist{}

	check := u.RedisConfig.Exists(wl.TableName())
	if check.Val() == 0 {
		results = wl.CacheAll(u)
	} else {
		cacheValue := u.RedisConfig.HGetAll(wl.TableName())
		for _, val := range cacheValue.Val() {
			obj := HfWhitelist{}

			_ = json.Unmarshal([]byte(val), &obj)
			results = append(results, obj)
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Id < results[j].Id
	})

	return results
}

func (wl *HfWhitelist) CacheAll(u HfUser) []HfWhitelist {
	results := []HfWhitelist{}
	u.DB.Find(&results)

	u.RedisConfig.Del(wl.TableName())

	pipe := u.RedisConfig.Pipeline()
	for _, val := range results {
		pipe.HSet(wl.TableName(), util.ToString(val.Id), util.JsonEndCode(val))
	}
	_, _ = pipe.Exec()
	return results
}
