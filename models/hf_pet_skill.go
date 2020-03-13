package models

import (
	"GoLang/libraries/constants"
	"GoLang/libraries/util"
	"encoding/json"
	"fmt"
	"github.com/kataras/iris"
	"sort"
)

type HfPetSkill struct {
	Skill    uint16 `gorm:"primary_key"`
	Level    uint16 `gorm:"primary_key"`
	PetLevel uint16
	Material string
}

func (HfPetSkill) TableName() string {
	return "hf_pet_skill"
}

func (ps *HfPetSkill) GetMap() iris.Map {
	return iris.Map{
		"skill":     ps.Skill,
		"level":     ps.Level,
		"pet_level": ps.PetLevel,
		"material":  util.JsonDecodeArray(ps.Material),
	}
}

func (ps *HfPetSkill) Find(skill uint16, level uint16, u HfUser) (HfPetSkill, bool) {
	ps.Skill = skill
	ps.Level = level
	field := fmt.Sprintf("%d_%d",  skill, level)

	count := 0
	cacheValue := u.RedisConfig.HGet(ps.TableName(), field)
	if cacheValue.Err() != nil {
		results := []HfPetSkill{};
		u.DB.Where(ps).First(&results).Count(&count)

		if count != 0 {
			*ps = results[0]
			u.RedisConfig.HSet(ps.TableName(), field, util.JsonEndCode(ps))
		}
	} else {
		count = 1
		_ = json.Unmarshal([]byte(cacheValue.Val()), &ps)
	}

	return *ps, count != 0
}

func (ps *HfPetSkill) CheckUpgradeStar(uPet HfUserPet, skillId uint16, user HfUser) (bool, string) {
	//uPet, pet là pet chủ

	material := util.JsonDecodeMap(ps.Material)

	//Level pet cần đạt
	if uPet.Level < ps.PetLevel {

		return false, `Level pet invalid`
	}

	//Gold yêu cầu
	if val, ok := material[constants.GOLD]; ok && user.Gold < uint(util.ToInt(val)) {

		return false, `Gold invalid`
	}

	//Đá tiến hóa
	if val, ok := material[constants.STONES]; ok {
		uStone := user.GetStones()
		stone := util.InterfaceToMap(val)
		for key, quan := range stone {
			if uStone[key] < uint(util.ToInt(quan)) {

				return false, `Stone ` + key + ` invalid`
			}
		}
	}

	return true, ``
}

func (ps *HfPetSkill) GetAll(u HfUser) []HfPetSkill {
	results := []HfPetSkill{}

	check := u.RedisConfig.Exists(ps.TableName())
	if check.Val() == 0 {
		results = ps.CacheAll(u)
	} else {
		cacheValue := u.RedisConfig.HGetAll(ps.TableName())
		for _, val := range cacheValue.Val() {
			obj := HfPetSkill{}

			_ = json.Unmarshal([]byte(val), &obj)
			results = append(results, obj)
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Skill < results[j].Skill || results[i].Skill == results[j].Skill && results[i].Level < results[j].Level
	})

	return results
}

func (ps *HfPetSkill) CacheAll(u HfUser) []HfPetSkill {
	results := []HfPetSkill{}
	u.DB.Find(&results)

	u.RedisConfig.Del(ps.TableName())

	pipe := u.RedisConfig.Pipeline()
	for _, val := range results {
		field := fmt.Sprintf("%d_%d", val.Skill, val.Level)
		pipe.HSet(ps.TableName(), field, util.JsonEndCode(val))
	}
	_, _ = pipe.Exec()
	return results
}
