package models

import (
	"GoLang/libraries/util"
	"encoding/json"
	"github.com/kataras/iris"
	"sort"
)

type HfSkill struct {
	Id           int `gorm:"primary_key"`
	Name         string
	ManaCost     int
	AngryCost    int
	PieceCost    int
	Cooldown     int
	Detail       string
	WeightAttack float32 `gorm:"default:0"`
	WeightDefend float32 `gorm:"default:0"`
	Type         uint16  `gorm:"default:0"`
	SkillConfig  string  `gorm:"default:{}"`
	Anim         string
	Tag          string
}

func (HfSkill) TableName() string {
	return "hf_skill"
}

func (sk *HfSkill) GetMap() iris.Map {
	return iris.Map{
		"id":            sk.Id,
		"name":          sk.Name,
		"mana_cost":     sk.ManaCost,
		"angry_cost":    sk.AngryCost,
		"piece_cost":    sk.PieceCost,
		"cooldown":      sk.Cooldown,
		"detail":        sk.Detail,
		"weight_attack": sk.WeightAttack,
		"weight_defend": sk.WeightDefend,
		"type":          sk.Type,
		"skill_config":  util.JsonDecodeMap(sk.SkillConfig),
		"anim":          sk.Anim,
		"tag":           sk.Tag,
	}
}

func (sk *HfSkill) GetAll(u HfUser) []HfSkill {
	results := []HfSkill{}

	check := u.RedisConfig.Exists(sk.TableName())
	if check.Val() == 0 {
		results = sk.CacheAll(u)
	} else {
		cacheValue := u.RedisConfig.HGetAll(sk.TableName())
		for _, val := range cacheValue.Val() {
			obj := HfSkill{}

			_ = json.Unmarshal([]byte(val), &obj)
			results = append(results, obj)
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Id < results[j].Id
	})

	return results
}

func (sk *HfSkill) CacheAll(u HfUser) []HfSkill {
	results := []HfSkill{}
	u.DB.Find(&results)

	u.RedisConfig.Del(sk.TableName())

	pipe := u.RedisConfig.Pipeline()
	for _, val := range results {
		pipe.HSet(sk.TableName(), util.ToString(val.Id), util.JsonEndCode(val))
	}
	_, _ = pipe.Exec()
	return results
}
