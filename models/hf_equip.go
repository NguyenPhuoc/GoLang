package models

import (
	"GoLang/libraries/util"
	"encoding/json"
	"github.com/kataras/iris"
	"sort"
)

type HfEquip struct {
	Id               uint16 `gorm:"primary_key"`
	NextId           uint16
	SetId            uint16
	Name             string
	Star             uint8
	Type             uint8
	Quality          uint8
	Damage           uint16
	Hp               uint16
	Armor            uint16
	ArmorPenetration uint16
	Critical         uint16
	ResistBranch     string
}

func (HfEquip) TableName() string {
	return "hf_equip"
}

func (e *HfEquip) Find(equipId uint16, u HfUser) (HfEquip, bool) {
	e.Id = equipId
	field := util.ToString(equipId)

	count := 0
	if equipId != 0 {
		cacheValue := u.RedisConfig.HGet(e.TableName(), field)
		if cacheValue.Err() != nil {
			u.DB.First(&e).Count(&count)

			if count != 0 {
				u.RedisConfig.HSet(e.TableName(), field, util.JsonEndCode(e))
			}
		} else {
			count = 1
			_ = json.Unmarshal([]byte(cacheValue.Val()), &e)
		}
	}

	return *e, count != 0
}

func (e *HfEquip) GetMap() iris.Map {
	return iris.Map{
		"id":                e.Id,
		"next_id":           e.NextId,
		"set_id":            e.SetId,
		"name":              e.Name,
		"star":              e.Star,
		"type":              e.Type,
		"quality":           e.Quality,
		"damage":            e.Damage,
		"hp":                e.Hp,
		"armor":             e.Armor,
		"armor_penetration": e.ArmorPenetration,
		"critical":          e.Critical,
		"resist_branch":     util.JsonDecodeMap(e.ResistBranch),
	}
}

func (e *HfEquip) GetAll(u HfUser) []HfEquip {
	results := []HfEquip{}

	check := u.RedisConfig.Exists(e.TableName())
	if check.Val() == 0 {
		results = e.CacheAll(u)
	} else {
		cacheValue := u.RedisConfig.HGetAll(e.TableName())
		for _, val := range cacheValue.Val() {
			obj := HfEquip{}

			_ = json.Unmarshal([]byte(val), &obj)
			results = append(results, obj)
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Id < results[j].Id
	})

	return results
}

func (e *HfEquip) CacheAll(u HfUser) []HfEquip {
	results := []HfEquip{}
	u.DB.Find(&results)

	u.RedisConfig.Del(e.TableName())

	pipe := u.RedisConfig.Pipeline()
	for _, val := range results {
		pipe.HSet(e.TableName(), util.ToString(val.Id), util.JsonEndCode(val))
	}
	_, _ = pipe.Exec()
	return results
}
