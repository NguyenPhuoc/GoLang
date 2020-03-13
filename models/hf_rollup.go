package models

import (
	"GoLang/libraries/util"
	"encoding/json"
	"fmt"
	"github.com/kataras/iris"
	"sort"
	"time"
)

type HfRollup struct {
	Month time.Time `gorm:"primary_key"`
	Gift  string
}

func (HfRollup) TableName() string {
	return "hf_rollup"
}

func (r *HfRollup) GetMap() iris.Map {
	return iris.Map{
		"month": r.Month,
		"gift":  util.JsonDecodeMap(r.Gift),
	}
}

func (r *HfRollup) Get(month time.Time, u HfUser) HfRollup {
	for {
		r.Month = month
		field := fmt.Sprintf("%d_%02d", month.Year(), month.Month())

		count := 0

		cacheValue := u.RedisConfig.HGet(r.TableName(), field)

		if cacheValue.Err() != nil {
			results := []HfRollup{};
			u.DB.Where(r).First(&results).Count(&count)

			if count != 0 {
				*r = results[0]
				u.RedisConfig.HSet(r.TableName(), field, util.JsonEndCode(r))
			} else {//lùi lại 1 tháng, tìm tháng trước đó
				month = month.AddDate(0, -1, 0)
				continue
			}
		} else {
			_ = json.Unmarshal([]byte(cacheValue.Val()), &r)
		}

		return *r
	}
}

func (r *HfRollup) GetAll(u HfUser) []HfRollup {
	results := []HfRollup{}

	check := u.RedisConfig.Exists(r.TableName())
	if check.Val() == 0 {
		results = r.CacheAll(u)
	} else {
		cacheValue := u.RedisConfig.HGetAll(r.TableName())
		for _, val := range cacheValue.Val() {
			obj := HfRollup{}

			_ = json.Unmarshal([]byte(val), &obj)
			results = append(results, obj)
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Month.Unix() < results[j].Month.Unix()
	})

	return results
}

func (r *HfRollup) CacheAll(u HfUser) []HfRollup {
	results := []HfRollup{}
	u.DB.Find(&results)

	u.RedisConfig.Del(r.TableName())

	pipe := u.RedisConfig.Pipeline()
	for _, val := range results {
		field := fmt.Sprintf("%d_%02d", val.Month.Year(), val.Month.Month())
		pipe.HSet(r.TableName(), field, util.JsonEndCode(val))
	}
	_, _ = pipe.Exec()
	return results
}
