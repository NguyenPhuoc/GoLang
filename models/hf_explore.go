package models

import (
	"GoLang/libraries/util"
	"encoding/json"
	"fmt"
	"github.com/kataras/iris"
	"sort"
)

type HfExplore struct {
	Node     uint `gorm:"primary_key"`
	MapId    uint `gorm:"primary_key"`
	Quantity float32
	Box      string
}

func (HfExplore) TableName() string {
	return "hf_explore"
}

func (e *HfExplore) GetMap() iris.Map {
	return iris.Map{
		"node":     e.Node,
		"map_id":   e.MapId,
		"quantity": e.Quantity,
		"box":      util.JsonDecodeMap(e.Box),
	}
}

func (e *HfExplore) GetExplore(node uint, mapId uint, u HfUser) (HfExplore, bool) {
	e.Node = node
	e.MapId = mapId
	field := fmt.Sprintf("%d_%d", mapId, node)

	count := 0
	if node != 0 && mapId != 0 {
		cacheValue := u.RedisConfig.HGet(e.TableName(), field)

		if cacheValue.Err() != nil {
			results := []HfExplore{};
			u.DB.Where(e).First(&results).Count(&count)

			if count != 0 {
				*e = results[0]
				u.RedisConfig.HSet(e.TableName(), field, util.JsonEndCode(e))
			}
		} else {
			count = 1
			_ = json.Unmarshal([]byte(cacheValue.Val()), &e)
		}
	}

	return *e, count != 0
}

func (e *HfExplore) GetAll(u HfUser) []HfExplore {
	results := []HfExplore{}

	check := u.RedisConfig.Exists(e.TableName())
	if check.Val() == 0 {
		results = e.CacheAll(u)
	} else {
		cacheValue := u.RedisConfig.HGetAll(e.TableName())
		for _, val := range cacheValue.Val() {
			obj := HfExplore{}

			_ = json.Unmarshal([]byte(val), &obj)
			results = append(results, obj)
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].MapId < results[j].MapId || results[i].MapId == results[j].MapId && results[i].Node < results[j].Node
	})

	return results
}

func (e *HfExplore) CacheAll(u HfUser) []HfExplore {
	results := []HfExplore{}
	u.DB.Find(&results)

	u.RedisConfig.Del(e.TableName())

	pipe := u.RedisConfig.Pipeline()
	for _, val := range results {
		field := fmt.Sprintf("%d_%d", val.MapId, val.Node)
		pipe.HSet(e.TableName(), field, util.JsonEndCode(val))
	}
	_, _ = pipe.Exec()
	return results
}
