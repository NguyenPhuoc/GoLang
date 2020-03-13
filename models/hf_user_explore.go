package models

import (
	"GoLang/libraries/constants"
	"GoLang/libraries/util"
	"encoding/json"
	"math"
	"sync"
	"time"
)

type HfUserExplore struct {
	UserId     string `gorm:"primary_key"`
	Node       uint
	MapId      uint
	StartTimes time.Time
	StartBox   time.Time
	Quantity   uint
}

func (HfUserExplore) TableName() string {
	return "hf_user_explore"
}

func (ue *HfUserExplore) UpdateCache(u HfUser) {
	u.RedisInfo.HSet(ue.UserId, ue.TableName(), util.JsonEndCode(ue))
}

func (ue *HfUserExplore) Save(u HfUser, wg *sync.WaitGroup) {
	defer wg.Done()
	ue.UpdateCache(u)
	u.DB.Save(&ue)
}

func (ue *HfUserExplore) Get(u HfUser) (HfUserExplore, map[string]interface{}, map[string]interface{}) {
	ue.UserId = u.UserId

	cacheValue := u.RedisInfo.HGet(ue.UserId, ue.TableName())
	if cacheValue.Err() == nil {

		_ = json.Unmarshal([]byte(cacheValue.Val()), &ue)
	} else {

		count := 0
		u.DB.Where(ue).Find(&ue).Count(&count)

		if count == 0 {
			ue.Node = 1
			ue.MapId = 1
			ue.StartTimes = time.Now()
			ue.StartBox = time.Now()

			var wg sync.WaitGroup
			wg.Add(1)
			go ue.Save(u, &wg)
			wg.Wait()
		} else {
			ue.UpdateCache(u)
		}
	}

	expl := HfExplore{}
	expl, checkExpl := expl.GetExplore(ue.Node, ue.MapId, u)

	if checkExpl {
		times := float32(math.Max(float64(time.Now().Unix()-ue.StartTimes.Unix()), 0))
		if maxTime := float32(3600 * 23); times > maxTime {
			times = maxTime
		}
		ue.Quantity += uint(times / 5 * expl.Quantity)
	}

	//Todo nhớ thay đổi bên config /config/explore
	conf := HfConfig{}
	stoneRatio, stoneEvoRatio, expRatio, goldRatio := conf.GetExploreRatio(u)

	stone := int(stoneRatio * float64(ue.Quantity))
	stone7 := int(stoneEvoRatio * float64(ue.Quantity))
	exp := int64(expRatio * float64(ue.Quantity))
	gold := int(goldRatio * float64(ue.Quantity))

	gifts := map[string]interface{}{
		constants.STONES: map[string]int{
			constants.FIRE:    stone,
			constants.EARTH:   stone,
			constants.THUNDER: stone,
			constants.WATER:   stone,
			constants.LIGHT:   stone,
			constants.DARK:    stone,
			constants.EVOLVE:  stone7,
		},
		constants.GOLD: gold,
		constants.EXP:  exp,
	}

	giftsReturn := map[string]interface{}{}
	if gold > 0 {
		giftsReturn[constants.GOLD] = gold
	}
	if exp > 0 {
		giftsReturn[constants.EXP] = exp
	}
	if stone > 0 || stone7 > 0 {
		_stone := map[string]int{}
		if stone > 0 {
			_stone[constants.FIRE] = stone
			_stone[constants.EARTH] = stone
			_stone[constants.THUNDER] = stone
			_stone[constants.WATER] = stone
			_stone[constants.LIGHT] = stone
			_stone[constants.DARK] = stone
		}
		if stone7 > 0 {
			_stone[constants.EVOLVE] = stone7
		}
		giftsReturn[constants.STONES] = _stone
	}

	return *ue, gifts, giftsReturn
}

func (ue HfUserExplore) GetQuantityBox() int {
	times := float32(math.Max(float64(time.Now().Unix()-ue.StartBox.Unix()), 0))
	if maxTime := float32(3600 * 23); times > maxTime {
		times = maxTime
	}
	quantityBox := int(times / (5 * 60))

	return quantityBox
}

func (ue *HfUserExplore) GetBox(u HfUser) interface{} {

	expl := HfExplore{}
	expl, checkExpl := expl.GetExplore(ue.Node, ue.MapId, u)

	quantityBox := ue.GetQuantityBox()

	if quantityBox > 0 && checkExpl{
		//expl.Box =`{"random":[{"gift":{"stones":{"f":1, "e":9},"exp":1,"gold":10,"piece_general":[{"type":1,"branch":"g","quantity":10}]},"percent":50},{"gift":{"gold":10,"piece_general":[{"type":2,"branch":"g","quantity":10}]},"percent":50}]}`
		confBox := util.JsonDecodeMap(expl.Box)

		gifts := map[string]interface{}{}
		if val, ok := confBox[constants.RANDOM]; ok {
			for i := 0; i < quantityBox; i++ {
				gift := util.RandomPercentGift(val)
				gifts = util.MergeGift(gifts, gift)
			}
		}

		return gifts
	}

	return nil
}
