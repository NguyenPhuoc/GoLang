package models

import (
	"GoLang/libraries/util"
	"encoding/json"
	"github.com/kataras/iris"
	"sort"
)

type HfPet struct {
	Id               uint16 `gorm:"primary_key"`
	Name             string
	Branch           string
	Type             uint8
	TypeRandom       uint8
	Rarity           uint
	Hp               int
	Damage           int
	Armor            float32
	ArmorPenetration float32
	Critical         float32
	ResistBranch     string
	Skill            string
	CanEnhance       int8
	CanOwned         int8
	Support          string
	Active           int8
}

func (HfPet) TableName() string {
	return "hf_pet"
}

func (p *HfPet) GetMap() iris.Map {
	return iris.Map{
		"id":                p.Id,
		"name":              p.Name,
		"branch":            p.Branch,
		"type":              p.Type,
		"type_random":       p.TypeRandom,
		"hp":                p.Hp,
		"damage":            p.Damage,
		"armor":             p.Armor,
		"armor_penetration": p.ArmorPenetration,
		"critical":          p.Critical,
		"resist_branch":     util.JsonDecodeMap(p.ResistBranch),
		"skill":             util.JsonDecodeMap(p.Skill),
		"can_enhance":       p.CanEnhance,
		"can_owned":         p.CanOwned,
		"support":           util.JsonDecodeMap(p.Support),
	}
}

func (p *HfPet) Find(petId uint16, u HfUser) (HfPet, bool) {
	p.Id = petId
	p.Active = 1
	field := util.ToString(petId)

	count := 0
	if petId != 0 {
		cacheValue := u.RedisConfig.HGet(p.TableName(), field)
		if cacheValue.Err() != nil {
			u.DB.First(&p).Count(&count)

			if count != 0 {
				u.RedisConfig.HSet(p.TableName(), field, util.JsonEndCode(p))
			}
		} else {
			count = 1
			_ = json.Unmarshal([]byte(cacheValue.Val()), &p)
		}
	}

	return *p, count != 0
}

func (p *HfPet) GetAll(u HfUser) []HfPet {
	results := []HfPet{}

	check := u.RedisConfig.Exists(p.TableName())
	if check.Val() == 0 {
		results = p.CacheAll(u)
	} else {
		cacheValue := u.RedisConfig.HGetAll(p.TableName())
		for _, val := range cacheValue.Val() {
			pet := HfPet{}

			_ = json.Unmarshal([]byte(val), &pet)
			results = append(results, pet)
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Id < results[j].Id
	})

	return results
}

func (p *HfPet) CacheAll(u HfUser) []HfPet {
	results := []HfPet{}
	u.DB.Where("active = 1").Find(&results)

	u.RedisConfig.Del(p.TableName())

	pipe := u.RedisConfig.Pipeline()
	for _, val := range results {
		pipe.HSet(p.TableName(), util.ToString(val.Id), util.JsonEndCode(val))
	}
	_, _ = pipe.Exec()
	return results
}
