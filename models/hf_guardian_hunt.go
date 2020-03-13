package models

import (
	"GoLang/libraries/util"
	"encoding/json"
	"fmt"
	"github.com/kataras/iris"
	"sort"
)

type HfGuardianHunt struct {
	GuardianId       uint16 `gorm:"primary_key"`
	Level            uint16 `gorm:"primary_key"`
	Hp               int
	Damage           int
	Critical         float64
	Armor            float64
	ArmorPenetration float64
	ResistBranch     string
	Iq               int
	LimitTurn        int
	TurnSkill        int
	SkillConfig      string
}

func (HfGuardianHunt) TableName() string {
	return "hf_guardian_hunt"
}

func (gh *HfGuardianHunt) GetMap() iris.Map {
	return iris.Map{
		"guardian_id":       gh.GuardianId,
		"level":             gh.Level,
		"hp":                gh.Hp,
		"damage":            gh.Damage,
		"critical":          gh.Critical,
		"armor":             gh.Armor,
		"armor_penetration": gh.ArmorPenetration,
		"resist_branch":     util.JsonDecodeMap(gh.ResistBranch),
		"iq":                gh.Iq,
		"limit_turn":        gh.LimitTurn,
		"turn_skill":        gh.TurnSkill,
		"skill_config":      util.JsonDecodeMap(gh.SkillConfig),
	}
}

func (gh *HfGuardianHunt) Find(guardianId uint16, level uint16, u HfUser) (HfGuardianHunt, bool) {
	gh.GuardianId = guardianId
	gh.Level = level
	field := fmt.Sprintf("%d_%d", gh.GuardianId, gh.Level)

	count := 0
	if guardianId != 0 && level != 0 {
		cacheValue := u.RedisConfig.HGet(gh.TableName(), field)

		if cacheValue.Err() != nil {
			u.DB.Where(gh).First(&gh).Count(&count)

			if count != 0 {
				u.RedisConfig.HSet(gh.TableName(), field, util.JsonEndCode(gh))
			}
		} else {
			count = 1
			_ = json.Unmarshal([]byte(cacheValue.Val()), &gh)
		}
	}

	return *gh, count != 0
}

func (gh *HfGuardianHunt) GetAll(u HfUser) []HfGuardianHunt {
	results := []HfGuardianHunt{}

	check := u.RedisConfig.Exists(gh.TableName())
	if check.Val() == 0 {
		results = gh.CacheAll(u)
	} else {
		cacheValue := u.RedisConfig.HGetAll(gh.TableName())
		for _, val := range cacheValue.Val() {
			obj := HfGuardianHunt{}

			_ = json.Unmarshal([]byte(val), &obj)
			results = append(results, obj)
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].GuardianId < results[j].GuardianId || results[i].GuardianId == results[j].GuardianId && results[i].Level < results[j].Level
	})

	return results
}

func (gh *HfGuardianHunt) CacheAll(u HfUser) []HfGuardianHunt {
	results := []HfGuardianHunt{}
	u.DB.Find(&results)

	u.RedisConfig.Del(gh.TableName())

	pipe := u.RedisConfig.Pipeline()
	for _, val := range results {
		field := fmt.Sprintf("%d_%d", val.GuardianId, val.Level)
		pipe.HSet(gh.TableName(), field, util.JsonEndCode(val))
	}
	_, _ = pipe.Exec()
	return results
}
