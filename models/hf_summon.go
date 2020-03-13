package models

import (
	"GoLang/libraries/util"
	"encoding/json"
	"github.com/kataras/iris"
	"sort"
)

type HfSummon struct {
	Id           uint8 `gorm:"primary_key"`
	Name         string
	BranchRandom string
	TypeRandom   string
	TimeFree     int
	TurnFee1     uint
	TurnFee10    uint
}

func (HfSummon) TableName() string {
	return "hf_summon"
}

func (s *HfSummon) GetMap() iris.Map {
	return iris.Map{
		"id":            s.Id,
		"name":          s.Name,
		"branch_random": s.BranchRandom,
		"type_random":   s.TypeRandom,
		"time_free":     s.TimeFree,
	}
}

func (s *HfSummon) Find(summonId uint8, u HfUser) (HfSummon, bool) {
	s.Id = summonId
	field := util.ToString(summonId)

	count := 0
	if summonId != 0 {
		cacheValue := u.RedisConfig.HGet(s.TableName(), field)
		if cacheValue.Err() != nil {
			u.DB.First(&s).Count(&count)

			if count != 0 {
				u.RedisConfig.HSet(s.TableName(), field, util.JsonEndCode(s))
			}
		} else {
			count = 1
			_ = json.Unmarshal([]byte(cacheValue.Val()), &s)
		}
	}

	return *s, count != 0
}

func (s *HfSummon) GetAll(u HfUser) []HfSummon {
	results := []HfSummon{}

	check := u.RedisConfig.Exists(s.TableName())
	if check.Val() == 0 {
		results = s.CacheAll(u)
	} else {
		cacheValue := u.RedisConfig.HGetAll(s.TableName())
		for _, val := range cacheValue.Val() {
			obj := HfSummon{}

			_ = json.Unmarshal([]byte(val), &obj)
			results = append(results, obj)
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Id < results[j].Id
	})

	return results
}

func (s *HfSummon) CacheAll(u HfUser) []HfSummon {
	results := []HfSummon{}
	u.DB.Find(&results)

	u.RedisConfig.Del(s.TableName())

	pipe := u.RedisConfig.Pipeline()
	for _, val := range results {
		pipe.HSet(s.TableName(), util.ToString(val.Id), util.JsonEndCode(val))
	}
	_, _ = pipe.Exec()
	return results
}
