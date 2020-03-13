package models

import (
	"GoLang/libraries/util"
	"encoding/json"
	"github.com/kataras/iris"
	"sort"
)

type HfLevel struct {
	Level uint `gorm:"primary_key"`
	Exp   uint64
	Gift  string
}

func (HfLevel) TableName() string {
	return "hf_level"
}

func (l *HfLevel) GetMap() iris.Map {
	return iris.Map{
		"level": l.Level,
		"exp":   l.Exp,
		"gift":  util.JsonDecodeMap(l.Gift),
	}
}

func (l *HfLevel) GetAll(u HfUser) []HfLevel {
	results := []HfLevel{}

	check := u.RedisConfig.Exists(l.TableName())
	if check.Val() == 0 {
		results = l.CacheAll(u)
	} else {
		cacheValue := u.RedisConfig.HGetAll(l.TableName())
		for _, val := range cacheValue.Val() {
			obj := HfLevel{}

			_ = json.Unmarshal([]byte(val), &obj)
			results = append(results, obj)
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Level < results[j].Level
	})

	return results
}

func (l *HfLevel) CacheAll(u HfUser) []HfLevel {
	results := []HfLevel{}
	u.DB.Find(&results)

	u.RedisConfig.Del(l.TableName())

	pipe := u.RedisConfig.Pipeline()
	for _, val := range results {
		pipe.HSet(l.TableName(), util.ToString(val.Level), util.JsonEndCode(val))
	}
	_, _ = pipe.Exec()
	return results
}

func (l *HfLevel) GetLevel(u HfUser) uint {
	levels := l.GetAll(u)
	for _, lv := range levels {
		if lv.Exp <= u.Exp {
			*l = lv
		} else {
			break
		}
	}
	return l.Level
}
