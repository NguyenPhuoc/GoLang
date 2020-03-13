package models

import (
	"GoLang/libraries/util"
	"github.com/jinzhu/gorm"
	"github.com/kataras/iris"
)

type HfTowerConfig struct {
	Floor            int `gorm:"primary_key"`
	BossId           int
	Armor            float64
	ArmorPenetration float64
	Critical         float64
	ResistBranch     string `gorm:"default:'{\"m\":0,\"w\":0,\"t\":0,\"f\":0,\"l\":0,\"d\":0}'"`
	Hp               int
	Damage           int
	Star             int
	Level            int
	Iq               int
	Support          string
	Guardian         string
	Gift             string
	Skill            string
}

func (HfTowerConfig) TableName() string {
	return "hf_tower_config"
}

func (tc *HfTowerConfig) GetMap() iris.Map {
	return iris.Map{
		"floor":             tc.Floor,
		"boss_id":           tc.BossId,
		"armor":             tc.Armor,
		"armor_penetration": tc.ArmorPenetration,
		"critical":          tc.Critical,
		"resist_branch":     util.JsonDecodeMap(tc.ResistBranch),
		"hp":                tc.Hp,
		"damage":            tc.Damage,
		"star":              tc.Star,
		"level":             tc.Level,
		"iq":                tc.Iq,
		"support":           util.JsonDecodeMap(tc.Support),
		"guardian":          util.JsonDecodeMap(tc.Guardian),
		"gift":              util.JsonDecodeMap(tc.Gift),
		"skill":             util.JsonDecodeMap(tc.Skill),
	}
}

func (tc *HfTowerConfig) Find(floor int, db *gorm.DB) (HfTowerConfig, bool) {
	tc.Floor = floor

	count := 0
	results := []HfTowerConfig{}
	db.Where(tc).First(&results).Count(&count)
	if count != 0 {
		*tc = results[0]
	}

	return *tc, count != 0
}

func (HfTowerConfig) MaxFloor(db *gorm.DB) int {
	tc := HfTowerConfig{}
	db.Last(&tc)

	return tc.Floor
}
