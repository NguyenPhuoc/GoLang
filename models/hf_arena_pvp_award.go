package models

import (
	"GoLang/libraries/util"
	"encoding/json"
	"github.com/kataras/iris"
	"sort"
)

type HfArenaPvpAward struct {
	Rank    int `gorm:"primary_key"`
	DesRank string
	BoxWeek string
}

func (HfArenaPvpAward) TableName() string {
	return "hf_arena_pvp_award"
}

func (apa *HfArenaPvpAward) GetMap() iris.Map {
	return iris.Map{
		"rank":     apa.Rank,
		"des_rank": apa.DesRank,
		"box_week": util.JsonDecodeMap(apa.BoxWeek),
	}
}

func (apa *HfArenaPvpAward) GetAll(u HfUser) []HfArenaPvpAward {
	results := []HfArenaPvpAward{}

	check := u.RedisConfig.Exists(apa.TableName())
	if check.Val() == 0 {
		results = apa.CacheAll(u)
	} else {
		cacheValue := u.RedisConfig.HGetAll(apa.TableName())
		for _, val := range cacheValue.Val() {
			obj := HfArenaPvpAward{}

			_ = json.Unmarshal([]byte(val), &obj)
			results = append(results, obj)
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Rank > results[j].Rank
	})

	return results
}

func (apa *HfArenaPvpAward) CacheAll(u HfUser) []HfArenaPvpAward {
	results := []HfArenaPvpAward{}
	u.DB.Find(&results)

	u.RedisConfig.Del(apa.TableName())

	pipe := u.RedisConfig.Pipeline()
	for _, val := range results {
		pipe.HSet(apa.TableName(), util.ToString(val.Rank), util.JsonEndCode(val))
	}
	_, _ = pipe.Exec()
	return results
}
