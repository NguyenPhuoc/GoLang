package models

import (
	"GoLang/libraries/constants"
	"GoLang/libraries/util"
	"encoding/json"
	"math/rand"
	"sort"
)

type HfShop struct {
	Id        int `gorm:"primary_key"`
	Name      string
	Price     uint
	TypePrice string
	Cost      uint
	Limit     int8
	IsRandom  uint8
	Ratio     uint
	Gift      string
}

func (HfShop) TableName() string {
	return "hf_shop"
}

func (s *HfShop) RandomMarketBlack(u HfUser, progress string) (idItems map[int]int, levelEquip int, progressBonus map[string]interface{}) {

	progressBonus = map[string]interface{}{}
	progressBonus = util.JsonDecodeMap(progress)

	itemBought := map[int]int8{}
	itemFree := map[int]int8{}
	util.JsonDecodeObject(util.JsonEndCode(progressBonus["limit"]), &itemBought)
	util.JsonDecodeObject(util.JsonEndCode(progressBonus["free"]), &itemFree)

	idItems = map[int]int{}
	ranFree := true//có mon free xuất hiện
	//nếu đã random trong ngày
	if len(itemFree) > 0 {
		ranFree = false

		//nếu chưa mua thì giữ nguyên
		for id, v := range itemFree {
			if v == 0 {
				idItems[1] = id
				break
			}
		}
	}

	for {
		allShop := s.GetAll(u)
		randomShop := []HfShop{}
		totalRatio := 0

		for _, shop := range allShop {
			if ranFree == false && shop.IsRandom == 1 || ranFree == true && shop.IsRandom == 2{
				//chưa mua, không giới hạn, đã mua nhưng chưa hết giới hạn
				if val, ok := itemBought[shop.Id]; !ok || shop.Limit == -1 || ok && val < shop.Limit {
					randomShop = append(randomShop, shop)

					totalRatio += int(shop.Ratio)
				}
			}
		}
		ratio := rand.Intn(totalRatio)
		per := 0

		for _, shop := range randomShop {
			per += int(shop.Ratio)
			if per > ratio {
				itemBought[shop.Id]++

				idItems[len(idItems)+1] = shop.Id
				if ranFree == true {
					ranFree = false
					progressBonus["free"] = map[int]int{shop.Id: 0}
				}

				break
			}
		}

		if len(idItems) == 10 {
			break
		}
	}

	//Check level equip
	levelEquip = 0
	uExplore := HfUserExplore{}
	uExplore.Get(u)

	expl := HfExplore{}
	expl, checkExpl := expl.GetExplore(uExplore.Node, uExplore.MapId, u)
	if checkExpl {
		confBox := util.JsonDecodeMap(expl.Box)
		if val, ok := confBox[constants.RANDOM]; ok {
			vals := util.InterfaceToArray(val)
			for _, val := range vals {
				val := util.InterfaceToMap(val)
				if val, ok := val[constants.RAND]; ok {
					item := util.InterfaceToMap(val)
					if val, ok := item[constants.EQUIP]; ok {
						levelEquip = util.ToInt(util.InterfaceToMap(val)[constants.LEVEL])
					}
				}
			}
		}
	}

	return
}

func (s *HfShop) Find(id int, u HfUser) (HfShop, bool) {
	s.Id = id

	cacheValue := u.RedisConfig.HGet(s.TableName(), util.ToString(id))

	count := 0
	if cacheValue.Err() != nil {
		u.DB.Where(s).First(&s).Count(&count)

		if count != 0 {
			u.RedisConfig.HSet(s.TableName(), util.ToString(id), util.JsonEndCode(s))
		}
	} else {
		count = 1
		_ = json.Unmarshal([]byte(cacheValue.Val()), &s)
	}

	return *s, count != 0
}

func (s *HfShop) GetAll(u HfUser) []HfShop {
	results := []HfShop{}

	check := u.RedisConfig.Exists(s.TableName())
	if check.Val() == 0 {
		results = s.CacheAll(u)
	} else {
		cacheValue := u.RedisConfig.HGetAll(s.TableName())
		for _, val := range cacheValue.Val() {
			obj := HfShop{}

			_ = json.Unmarshal([]byte(val), &obj)
			results = append(results, obj)
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Id < results[j].Id
	})

	return results
}

func (s *HfShop) CacheAll(u HfUser) []HfShop {
	results := []HfShop{}
	u.DB.Find(&results)

	u.RedisConfig.Del(s.TableName())

	pipe := u.RedisConfig.Pipeline()
	for _, val := range results {
		if !util.InArray(val.Id, []int{34, 35, 36}) {
			pipe.HSet(s.TableName(), util.ToString(val.Id), util.JsonEndCode(val))
		}
	}
	_, _ = pipe.Exec()
	return results
}
