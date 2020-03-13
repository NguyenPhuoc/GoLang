package models

import (
	"GoLang/libraries/util"
	"encoding/json"
	"fmt"
	"sort"
)

type HfConfig struct {
	Key   string `gorm:"primary_key" json:"key"`
	Value string `json:"value"`
	Note  string `json:"note"`
}

func (HfConfig) TableName() string {
	return "hf_config"
}

func (conf *HfConfig) Find(key string, u HfUser) HfConfig {
	conf.Key = key

	count := 0
	if key != "" {
		cacheValue := u.RedisConfig.HGet(conf.TableName(), key)
		if cacheValue.Err() != nil {
			u.DB.First(&conf).Count(&count)

			if count != 0 {
				u.RedisConfig.HSet(conf.TableName(), key, util.JsonEndCode(conf))
			}
		} else {
			count = 1
			_ = json.Unmarshal([]byte(cacheValue.Val()), &conf)
		}
	}

	if count == 0 {
		fmt.Println("hf_config" + "." + key + " not Exists")
	}

	return *conf
}

func (conf *HfConfig) GetAll(u HfUser) []HfConfig {
	results := []HfConfig{}

	check := u.RedisConfig.Exists(conf.TableName())
	if check.Val() == 0 {
		results = conf.CacheAll(u)
	} else {
		cacheValue := u.RedisConfig.HGetAll(conf.TableName())
		for _, val := range cacheValue.Val() {
			obj := HfConfig{}

			_ = json.Unmarshal([]byte(val), &obj)
			results = append(results, obj)
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Key < results[j].Key
	})

	return results
}

func (conf *HfConfig) CacheAll(u HfUser) []HfConfig {
	results := []HfConfig{}
	u.DB.Find(&results)

	u.RedisConfig.Del(conf.TableName())

	pipe := u.RedisConfig.Pipeline()
	for _, val := range results {
		pipe.HSet(conf.TableName(), val.Key, util.JsonEndCode(val))
	}
	_, _ = pipe.Exec()
	return results
}

func (conf *HfConfig) GetExploreRatio(u HfUser) (stoneRatio, stoneEvoRatio, expRatio, goldRatio float64) {

	stoneRatioKey, stoneEvoRatioKey, expRatioKey, goldRatioKey := "explore_stone_ratio", "explore_stone_evo_ratio", "explore_exp_ratio", "explore_gold_ratio"

	con := HfConfig{}
	stoneRatio = util.ToFloat(con.Find(stoneRatioKey, u).Value)

	con = HfConfig{}
	stoneEvoRatio = util.ToFloat(con.Find(stoneEvoRatioKey, u).Value)

	con = HfConfig{}
	expRatio = util.ToFloat(con.Find(expRatioKey, u).Value)

	con = HfConfig{}
	goldRatio = util.ToFloat(con.Find(goldRatioKey, u).Value)
	return
}

func (conf *HfConfig) GetPackageGift(packageId string, u HfUser) (giftGem, goldActive int) {

	giftGemKey, goldActiveKey := packageId+"_gift", packageId+"_active"

	con := HfConfig{}
	giftGem = util.ToInt(con.Find(giftGemKey, u).Value)

	con = HfConfig{}
	goldActive = util.ToInt(con.Find(goldActiveKey, u).Value)
	return
}
