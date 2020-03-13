package models

import (
	"GoLang/libraries/util"
	"encoding/json"
	"fmt"
	"github.com/kataras/iris"
	"sort"
)

type HfCampaign struct {
	Node             uint `gorm:"primary_key"`
	MapId            uint `gorm:"primary_key"`
	MapDiff          uint `gorm:"primary_key"`
	BossId           uint
	Armor            float32
	ArmorPenetration float32
	Critical         float32
	ResistBranch     string
	Hp               int
	Damage           int
	Star             int
	Level            int
	Iq               int
	LimitTurn        int
	Gift             string
	IsBoss           uint8
	RequireLevel     uint16
	Support          string
	Background       int
	Guardian         string
	Skill            string
}

func (HfCampaign) TableName() string {
	return "hf_campaign"
}

func (cam *HfCampaign) GetMap() iris.Map {
	return iris.Map{
		"node":              cam.Node,
		"map_id":            cam.MapId,
		"map_diff":          cam.MapDiff,
		"boss_id":           cam.BossId,
		"armor":             cam.Armor,
		"armor_penetration": cam.ArmorPenetration,
		"critical":          cam.Critical,
		"resist_branch":     util.JsonDecodeMap(cam.ResistBranch),
		"hp":                cam.Hp,
		"damage":            cam.Damage,
		"star":              cam.Star,
		"level":             cam.Level,
		"iq":                cam.Iq,
		"limit_turn":        cam.LimitTurn,
		"gift":              util.JsonDecodeMap(cam.Gift),
		"is_boss":           cam.IsBoss,
		"require_level":     cam.RequireLevel,
		"support":           util.JsonDecodeMap(cam.Support),
		"background":        cam.Background,
		"guardian":          util.JsonDecodeMap(cam.Guardian),
		"skill":             util.JsonDecodeMap(cam.Skill),
	}
}

func (cam *HfCampaign) Find(mapDiff, mapId, node uint, u HfUser) HfCampaign {
	cam.Node = node
	cam.MapId = mapId
	cam.MapDiff = mapDiff

	field := fmt.Sprintf("%d_%d_%d", mapDiff, mapId, node)
	cacheValue := u.RedisConfig.HGet(cam.TableName(), field)

	count := 0
	if cacheValue.Err() != nil {
		u.DB.Where(cam).First(&cam).Count(&count)

		if count != 0 {
			u.RedisConfig.HSet(cam.TableName(), field, util.JsonEndCode(cam))
		}
	} else {
		_ = json.Unmarshal([]byte(cacheValue.Val()), &cam)
	}

	return *cam
}

func (cam *HfCampaign) FindLastMap(mapDiff, mapId uint, u HfUser) HfCampaign {
	all := cam.GetAll(u)

	for _, val := range all {
		if val.MapDiff == mapDiff && val.MapId == mapId {
			*cam = val
		}
	}

	return *cam
}

func (cam *HfCampaign) FindLastDiff(mapDiff uint, u HfUser) HfCampaign {
	all := cam.GetAll(u)

	for _, val := range all {
		if val.MapDiff == mapDiff {
			*cam = val
		}
	}

	return *cam
}

func (cam *HfCampaign) FindMaxDiff(u HfUser) HfCampaign {
	all := cam.GetAll(u)
	*cam = all[len(all)-1]

	return *cam
}

func (cam *HfCampaign) FindMaxNodeMap(u HfUser) HfCampaign {
	all := cam.GetAll(u)
	*cam = all[len(all)-1]

	return *cam
}

func (cam *HfCampaign) GetAll(u HfUser) []HfCampaign {
	results := []HfCampaign{}

	check := u.RedisConfig.Exists(cam.TableName())
	if check.Val() == 0 {
		results = cam.CacheAll(u)
	} else {
		cacheValue := u.RedisConfig.HGetAll(cam.TableName())
		for _, val := range cacheValue.Val() {
			c := HfCampaign{}

			_ = json.Unmarshal([]byte(val), &c)
			results = append(results, c)
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].MapDiff < results[j].MapDiff || results[i].MapDiff == results[j].MapDiff && results[i].MapId < results[j].MapId || results[i].MapDiff == results[j].MapDiff && results[i].MapId == results[j].MapId && results[i].Node < results[j].Node
	})

	return results
}

func (cam *HfCampaign) CacheAll(u HfUser) []HfCampaign {
	results := []HfCampaign{}
	u.DB.Find(&results)

	u.RedisConfig.Del(cam.TableName())

	pipe := u.RedisConfig.Pipeline()
	for _, val := range results {
		field := fmt.Sprintf("%d_%d_%d", val.MapDiff, val.MapId, val.Node)
		pipe.HSet(cam.TableName(), field, util.JsonEndCode(val))
	}
	_, _ = pipe.Exec()
	return results
}
