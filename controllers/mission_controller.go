package controllers

import (
	"GoLang/libraries/constants"
	"GoLang/libraries/logtype"
	"GoLang/libraries/util"
	"GoLang/models"
	"github.com/kataras/iris"
	"sort"
	"sync"
	"time"
)

type MissionController struct {
	MyController
}

// /mission/daily/info
func (c *MissionController) GetDailyInfo() {
	user := c.User

	uMission := models.HfUserMission{}
	uMission.GetDaily(user)

	mission := models.HfMission{}
	missionDaily, _, _ := mission.GetProgressDaily(user)
	bigGift := mission.GetBigGiftDaily(user)

	progressMission := map[int][]int{}
	progressGift := map[int]int{}

	util.JsonDecodeObject(uMission.Progress, &progressMission)
	util.JsonDecodeObject(uMission.Gift, &progressGift)

	dataMission := []iris.Map{}

	var missionIds []int
	for k, _ := range progressMission {
		missionIds = append(missionIds, k)
	}
	sort.Ints(missionIds)

	for _, missionId := range missionIds {
		for _, mis := range missionDaily {
			if missionId == mis.MissionId {

				dataMission = append(dataMission, iris.Map{
					"id":       mis.MissionId,
					"progress": progressMission[missionId],
					"status":   util.StatusGift(progressGift[missionId]),
					"key":      mis.MissionKey,
					"gift":     util.JsonDecodeMap(mis.Gift),
				})

				break
			}
		}
	}

	timeReset := uMission.DailyNextReset().Unix() - time.Now().Unix()

	c.DataResponse = iris.Map{"code": 1, "data": iris.Map{
		"missions":   dataMission,
		"status":     util.StatusGift(uMission.BigGift),
		"big_gift":   bigGift,
		"time_reset": timeReset,
	}}
}

// /mission/daily/complete
func (c *MissionController) PostDailyComplete(form formValue, ctx iris.Context) {
	if c.validToken(ctx) {
		user := c.User
		missionId := util.ParseInt(form("mission_id"))

		uMission := models.HfUserMission{}
		uMission.GetDaily(user)

		uMission.CompleteDaily(missionId, user, 1)

		uMission.CheckBigGiftDaily(user)

		c.DataResponse = iris.Map{"code": 1, "msg": "Success"}
	}
}

// /mission/daily/gift
func (c *MissionController) PostDailyGift(form formValue, ctx iris.Context) {
	if c.validToken(ctx) {
		user := c.User
		missionId := util.ParseInt(form("mission_id"))

		uMission := models.HfUserMission{}
		uMission.GetDaily(user)

		progressGift := map[int]int{}
		util.JsonDecodeObject(uMission.Gift, &progressGift)

		if status, ok := progressGift[missionId]; !ok || status != 0 {
			c.DataResponse = iris.Map{"code": -1, "msg": "Mission or gift invalid"}
			return
		}

		mission := models.HfMission{}
		gifts := mission.GetGiftDaily(missionId, user)

		//update status, gift
		progressGift[missionId] = 1
		uMission.Gift = util.JsonEndCode(progressGift)

		var wg sync.WaitGroup
		wg.Add(2)
		go uMission.Save(user, &wg)
		go user.UpdateGifts(gifts, logtype.GIFT_MISSION_DAILY, 0, &wg)
		wg.Wait()

		c.DataResponse = iris.Map{"code": 1, "msg": "Success", "data": iris.Map{constants.GIFT: gifts}}
	}
}

// /mission/daily/biggift
func (c *MissionController) PostDailyBiggift(form formValue, ctx iris.Context) {
	if c.validToken(ctx) {
		user := c.User

		uMission := models.HfUserMission{}
		uMission.GetDaily(user)

		progressGift := map[int]int{}
		util.JsonDecodeObject(uMission.Gift, &progressGift)

		if uMission.BigGift != 0 {
			c.DataResponse = iris.Map{"code": -1, "msg": "Big gift invalid"}
			return
		}

		mission := models.HfMission{}
		gifts := mission.GetBigGiftDaily(user)

		//update status, gift
		uMission.BigGift = 1

		var wg sync.WaitGroup
		wg.Add(2)
		go uMission.Save(user, &wg)
		go user.UpdateGifts(gifts, logtype.GIFT_BIG_MISSION_DAILY, 0, &wg)
		wg.Wait()

		c.DataResponse = iris.Map{"code": 1, "msg": "Success", "data": iris.Map{constants.GIFT: gifts}}
	}
}
