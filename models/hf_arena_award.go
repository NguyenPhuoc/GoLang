package models

import (
	"GoLang/libraries/util"
	"encoding/json"
	"github.com/kataras/iris"
	"sort"
)

type HfArenaAward struct {
	Rank     int `gorm:"primary_key"`
	DesRank  string
	BoxDay   string
	BoxWeek  string
	BoxMonth string
}

func (HfArenaAward) TableName() string {
	return "hf_arena_award"
}

func (aa *HfArenaAward) GetMap() iris.Map {
	return iris.Map{
		"rank":      aa.Rank,
		"des_rank":  aa.DesRank,
		"box_day":   util.JsonDecodeMap(aa.BoxDay),
		"box_week":  util.JsonDecodeMap(aa.BoxWeek),
		"box_month": util.JsonDecodeMap(aa.BoxMonth),
	}
}

func (aa *HfArenaAward) GetAll(u HfUser) []HfArenaAward {
	results := []HfArenaAward{}

	check := u.RedisConfig.Exists(aa.TableName())
	if check.Val() == 0 {
		results = aa.CacheAll(u)
	} else {
		cacheValue := u.RedisConfig.HGetAll(aa.TableName())
		for _, val := range cacheValue.Val() {
			obj := HfArenaAward{}

			_ = json.Unmarshal([]byte(val), &obj)
			results = append(results, obj)
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Rank > results[j].Rank
	})

	return results
}

func (aa *HfArenaAward) CacheAll(u HfUser) []HfArenaAward {
	results := []HfArenaAward{}
	u.DB.Find(&results)

	u.RedisConfig.Del(aa.TableName())

	pipe := u.RedisConfig.Pipeline()
	for _, val := range results {
		pipe.HSet(aa.TableName(), util.ToString(val.Rank), util.JsonEndCode(val))
	}
	_, _ = pipe.Exec()
	return results
}
