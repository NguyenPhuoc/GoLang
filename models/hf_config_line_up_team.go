package models

import (
	"GoLang/libraries/util"
	"encoding/json"
	"github.com/kataras/iris"
	"sort"
)

type HfConfigLineUpTeam struct {
	Id           string `gorm:"primary_key"`
	Damage       float64
	Hp           float64
	Critical     float64
	ResistBranch float64
	Armor        float64
}

func (HfConfigLineUpTeam) TableName() string {
	return "hf_config_line_up_team"
}

func (conf *HfConfigLineUpTeam) GetMap() iris.Map {
	return iris.Map{
		"id":            conf.Id,
		"damage":        conf.Damage,
		"hp":            conf.Hp,
		"critical":      conf.Critical,
		"resist_branch": conf.ResistBranch,
		"armor":         conf.Armor,
	}
}

func (conf *HfConfigLineUpTeam) GetAll(u HfUser) []HfConfigLineUpTeam {
	results := []HfConfigLineUpTeam{}

	check := u.RedisConfig.Exists(conf.TableName())
	if check.Val() == 0 {
		results = conf.CacheAll(u)
	} else {
		cacheValue := u.RedisConfig.HGetAll(conf.TableName())
		for _, val := range cacheValue.Val() {
			obj := HfConfigLineUpTeam{}

			_ = json.Unmarshal([]byte(val), &obj)
			results = append(results, obj)
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Id < results[j].Id
	})

	return results
}

func (conf *HfConfigLineUpTeam) CacheAll(u HfUser) []HfConfigLineUpTeam {
	results := []HfConfigLineUpTeam{}
	u.DB.Find(&results)

	u.RedisConfig.Del(conf.TableName())

	pipe := u.RedisConfig.Pipeline()
	for _, val := range results {
		pipe.HSet(conf.TableName(), val.Id, util.JsonEndCode(val))
	}
	_, _ = pipe.Exec()
	return results
}
