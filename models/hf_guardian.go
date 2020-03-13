package models

import (
	"GoLang/libraries/util"
	"encoding/json"
	"github.com/kataras/iris"
	"sort"
)

type HfGuardian struct {
	Id                     uint16 `gorm:"primary_key"`
	Name                   string
	Branch                 string
	ConfigHp               string
	ConfigHpEvolve         string
	ConfigDamage           string
	ConfigDamageEvolve     string
	ConfigCritical         string
	ConfigArmorPenetration string
	ConfigArmor            string
	ConfigResistBranch     string
	ConfigBonusSkill       string
	ConfigActiveSkill      string
	Cost                   int
}

func (HfGuardian) TableName() string {
	return "hf_guardian"
}

func (g *HfGuardian) GetMap() iris.Map {
	return iris.Map{
		"id":                       g.Id,
		"name":                     g.Name,
		"branch":                   g.Branch,
		"config_hp":                util.JsonDecodeArray(g.ConfigHp),
		"config_hp_evolve":         util.JsonDecodeArray(g.ConfigHpEvolve),
		"config_damage":            util.JsonDecodeArray(g.ConfigDamage),
		"config_damage_evolve":     util.JsonDecodeArray(g.ConfigDamageEvolve),
		"config_critical":          util.JsonDecodeArray(g.ConfigCritical),
		"config_armor_penetration": util.JsonDecodeArray(g.ConfigArmorPenetration),
		"config_armor":             util.JsonDecodeArray(g.ConfigArmor),
		"config_resist_branch":     util.JsonDecodeArray(g.ConfigResistBranch),
		"config_bonus_skill":       util.JsonDecodeArray(g.ConfigBonusSkill),
		"config_active_skill":      util.JsonDecodeArray(g.ConfigActiveSkill),
		"cost":                     g.Cost,
	}
}

func (g *HfGuardian) Find(id uint16, u HfUser) (HfGuardian, bool) {
	g.Id = id

	cacheValue := u.RedisConfig.HGet(g.TableName(), util.ToString(id))

	count := 0
	if cacheValue.Err() != nil {
		u.DB.Where(g).First(&g).Count(&count)

		if count != 0 {
			u.RedisConfig.HSet(g.TableName(), util.ToString(id), util.JsonEndCode(g))
		}
	} else {
		count = 1
		_ = json.Unmarshal([]byte(cacheValue.Val()), &g)
	}

	return *g, count != 0
}

func (g *HfGuardian) GetAll(u HfUser) []HfGuardian {
	results := []HfGuardian{}

	check := u.RedisConfig.Exists(g.TableName())
	if check.Val() == 0 {
		results = g.CacheAll(u)
	} else {
		cacheValue := u.RedisConfig.HGetAll(g.TableName())
		for _, val := range cacheValue.Val() {
			obj := HfGuardian{}

			_ = json.Unmarshal([]byte(val), &obj)
			results = append(results, obj)
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Id < results[j].Id
	})

	return results
}

func (g *HfGuardian) CacheAll(u HfUser) []HfGuardian {
	results := []HfGuardian{}
	u.DB.Find(&results)

	u.RedisConfig.Del(g.TableName())

	pipe := u.RedisConfig.Pipeline()
	for _, val := range results {
		pipe.HSet(g.TableName(), util.ToString(val.Id), util.JsonEndCode(val))
	}
	_, _ = pipe.Exec()
	return results
}
