package models

import (
	"GoLang/libraries/util"
	"encoding/json"
	"sync"
	"time"
)

const (
	QDL_CAMPAIGN     = 1
	QDL_SUMMON_1     = 2
	QDL_SUMMON_2     = 3
	QDL_ARENA_PVE    = 4
	QDL_ARENA_PVP    = 5
	QDL_BUY_GOLD     = 6
	QDL_FORGE        = 7
	QDL_HUNTGUARDIAN = 8
	QDL_MARKET_GOLD  = 9
	QDL_MARKET_GEM   = 10
)

type HfUserMission struct {
	UserId    string `gorm:"primary_key"`
	Key       string `gorm:"primary_key"`
	Progress  string
	Gift      string
	BigGift   uint8
	TimeReset time.Time
}

func (HfUserMission) TableName() string {
	return "hf_user_mission"
}

func (um *HfUserMission) UpdateCache(u HfUser) {
	u.RedisInfo.HSet(um.UserId, um.TableName(), util.JsonEndCode(um))
}

func (um *HfUserMission) Save(u HfUser, wg *sync.WaitGroup) {
	defer wg.Done()
	um.UpdateCache(u)
	u.DB.Save(&um)
}

func (um *HfUserMission) Get(key string, u HfUser) HfUserMission {
	um.Key = key
	um.UserId = u.UserId

	mission := HfMission{}

	cacheValue := u.RedisInfo.HGet(um.UserId, um.TableName())
	if cacheValue.Err() == nil {

		_ = json.Unmarshal([]byte(cacheValue.Val()), &um)
	} else {

		count := 0
		results := []HfUserMission{};
		u.DB.Where(um).First(&results).Count(&count)

		if count == 0 {
			_, progressMission, progressGift := mission.GetProgressDaily(u)

			um.Progress = util.JsonEndCode(progressMission)
			um.Gift = util.JsonEndCode(progressGift)
			um.BigGift = 2
			um.TimeReset = um.DailyLastReset()

			var wg sync.WaitGroup
			wg.Add(1)
			go um.Save(u, &wg)
			wg.Wait()
		} else {
			*um = results[0]
			um.UpdateCache(u)
		}
	}

	//Todo reset lại nv hàng ngày
	if um.TimeReset.Unix() < um.DailyLastReset().Unix() {
		_, progressMission, progressGift := mission.GetProgressDaily(u)

		um.Progress = util.JsonEndCode(progressMission)
		um.Gift = util.JsonEndCode(progressGift)
		um.BigGift = 2
		um.TimeReset = um.DailyLastReset()

		var wg sync.WaitGroup
		wg.Add(1)
		go um.Save(u, &wg)
		wg.Wait()
	}

	um.CheckBigGiftDaily(u)

	return *um
}

func (um *HfUserMission) GetDaily(u HfUser) HfUserMission {

	return um.Get("daily", u)
}

func (um *HfUserMission) CompleteDaily(missionId int, u HfUser, quantity int) {

	progressMission := map[int][]int{}
	progressGift := map[int]int{}

	util.JsonDecodeObject(um.Progress, &progressMission)
	util.JsonDecodeObject(um.Gift, &progressGift)

	if progress, ok := progressMission[missionId]; ok {
		if progress[0] < progress[1] && quantity > 0 {
			progress[0] += quantity
			if progress[0] > progress[1] {
				progress[0] = progress[1]
			}
			progressMission[missionId] = progress

			//Hoàn thành thì có quà
			if progress[0] >= progress[1] && progressGift[missionId] == 2 {
				progressGift[missionId] = 0
			}
		}

		um.Progress = util.JsonEndCode(progressMission)
		um.Gift = util.JsonEndCode(progressGift)

		var wg sync.WaitGroup
		wg.Add(1)
		go um.Save(u, &wg)
		wg.Wait()
	}
}

func (um *HfUserMission) CheckBigGiftDaily(u HfUser) {

	if um.BigGift == 2 {
		progressMission := map[int][]int{}
		util.JsonDecodeObject(um.Progress, &progressMission)

		mission := HfMission{}
		missionDaily, _, _ := mission.GetProgressDaily(u)

		//có quà bự không
		check := true
		for misId, progress := range progressMission {
			for _, mis := range missionDaily {
				if mis.MissionId == misId {//nếu tồn tại nv mới check
					if progress[0] < progress[1] {

						check = false
					}
					break
				}
			}
			//nếu không đủ 1 đk nào đó thì không cần check nữa
			if check == false {
				break
			}
		}

		if check {
			um.BigGift = 0
			var wg sync.WaitGroup
			wg.Add(1)
			go um.Save(u, &wg)
			wg.Wait()
		}
	}
}

func (um *HfUserMission) DailyNextReset() time.Time {
	t := time.Now()
	timeCheck := time.Date(t.Year(), t.Month(), t.Day(), 5, 0, 0, 0, time.Local)
	if timeCheck.Unix() < t.Unix() {
		timeCheck = timeCheck.AddDate(0, 0, 1)
	}
	return timeCheck
}

func (um *HfUserMission) DailyLastReset() time.Time {
	t := time.Now()
	timeCheck := time.Date(t.Year(), t.Month(), t.Day(), 5, 0, 0, 0, time.Local)
	if timeCheck.Unix() > t.Unix() {
		timeCheck = timeCheck.AddDate(0, 0, -1)
	}
	return timeCheck
}
