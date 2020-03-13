package models

import (
	"GoLang/libraries/util"
	"encoding/json"
	"github.com/kataras/iris"
	"sort"
)

type HfCampaignInfo struct {
	MapId   uint `gorm:"primary_key"`
	MapName string
}

func (HfCampaignInfo) TableName() string {
	return "hf_campaign_info"
}

func (cam *HfCampaignInfo) GetMap() iris.Map {
	return iris.Map{
		"map_id":   cam.MapId,
		"map_name": cam.MapName,
	}
}

func (cam *HfCampaignInfo) GetAll(u HfUser) []HfCampaignInfo {
	results := []HfCampaignInfo{}

	check := u.RedisConfig.Exists(cam.TableName())
	if check.Val() == 0 {
		results = cam.CacheAll(u)
	} else {
		cacheValue := u.RedisConfig.HGetAll(cam.TableName())
		for _, val := range cacheValue.Val() {
			c := HfCampaignInfo{}

			_ = json.Unmarshal([]byte(val), &c)
			results = append(results, c)
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].MapId < results[j].MapId
	})

	return results
}

func (cam *HfCampaignInfo) CacheAll(u HfUser) []HfCampaignInfo {
	results := []HfCampaignInfo{}
	u.DB.Find(&results)

	u.RedisConfig.Del(cam.TableName())

	pipe := u.RedisConfig.Pipeline()
	for _, val := range results {
		pipe.HSet(cam.TableName(), util.ToString(val.MapId), util.JsonEndCode(val))
	}
	_, _ = pipe.Exec()
	return results
}
