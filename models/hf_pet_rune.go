package models

import (
	"GoLang/libraries/util"
	"database/sql"
	"encoding/json"
	"sort"
)

type HfPetRune struct {
	Type     string `gorm:"primary_key"`
	Name     sql.NullString
	MaxLevel int
	Material float64
}

func (HfPetRune) TableName() string {
	return "hf_pet_rune"
}

func (pr *HfPetRune) Find(typeRune string, u HfUser) (HfPetRune, bool) {
	pr.Type = typeRune
	field := pr.Type

	count := 0
	if typeRune != "" {
		cacheValue := u.RedisConfig.HGet(pr.TableName(), field)
		if cacheValue.Err() != nil {
			results := []HfPetRune{};
			u.DB.Where(pr).First(&results).Count(&count)

			if count != 0 {
				*pr = results[0]
				u.RedisConfig.HSet(pr.TableName(), field, util.JsonEndCode(pr))
			}
		} else {
			count = 1
			_ = json.Unmarshal([]byte(cacheValue.Val()), &pr)
		}
	}

	return *pr, count != 0
}

func (pr *HfPetRune) GetAll(u HfUser) []HfPetRune {
	results := []HfPetRune{}

	check := u.RedisConfig.Exists(pr.TableName())
	if check.Val() == 0 {
		results = pr.CacheAll(u)
	} else {
		cacheValue := u.RedisConfig.HGetAll(pr.TableName())
		for _, val := range cacheValue.Val() {
			obj := HfPetRune{}

			_ = json.Unmarshal([]byte(val), &obj)
			results = append(results, obj)
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Type < results[j].Type
	})

	return results
}

func (pr *HfPetRune) CacheAll(u HfUser) []HfPetRune {
	results := []HfPetRune{}
	u.DB.Find(&results)

	u.RedisConfig.Del(pr.TableName())

	pipe := u.RedisConfig.Pipeline()
	for _, val := range results {
		pipe.HSet(pr.TableName(), val.Type, util.JsonEndCode(val))
	}
	_, _ = pipe.Exec()
	return results
}
