package controllers

import (
	"GoLang/libraries/constants"
	"GoLang/libraries/logtype"
	"GoLang/libraries/util"
	"GoLang/models"
	"github.com/kataras/iris"
	"sort"
	"sync"
)

type NewbieController struct {
	MyController
}

// /newbie/update/step
func (c *NewbieController) PostUpdateStep(form formValue, ctx iris.Context) {
	if c.validToken(ctx) {
		user := c.User
		step := util.ParseInt(form("step"))

		var wg sync.WaitGroup
		wg.Add(1)
		go user.SetNewbieStep(step, &wg)
		wg.Wait()

		c.DataResponse = iris.Map{"code": 1, "msg": "Success"}
	}
}

// /newbie/gift
func (c *NewbieController) PostGift(form formValue, ctx iris.Context) {
	if c.validToken(ctx) {
		user := c.User
		step := util.ToInt(form("step"))

		uNewbie := models.HfUserNewbie{}
		uNewbie.Get(user)

		progress := map[int]int{}
		util.JsonDecodeObject(uNewbie.StepGift, &progress)

		mission := models.HfMissionNewbie{}

		switch step {
		case 1://Quà trận demo
			if val, ok := progress[step]; ok && val == 0 {
				gifts := mission.GetStepGift(step, user)

				var wg sync.WaitGroup
				wg.Add(2)
				progress[step] = 1
				uNewbie.StepGift = util.JsonEndCode(progress)
				go func() {
					c.DB.Save(&uNewbie)
					wg.Done()
				}()
				go user.UpdateGifts(gifts, logtype.GIFT_NEWBIE, 0, &wg)
				wg.Wait()

				c.DataResponse = iris.Map{"code": 1, "data": iris.Map{constants.GIFT: gifts}}
			} else {
				c.DataResponse = iris.Map{"code": -1, "msg": "Gift invalid"}
			}

		case 2://Show quà Thám hiểm
			gifts := mission.GetStepGift(step, user)

			c.DataResponse = iris.Map{"code": 1, "data": iris.Map{
				constants.GIFT: gifts,
				"next_time":    99999,
				"map_id":       1,
				"node":         1,
				"quantity_box": 1,
				"total_time":   90000,
			}}

		case 3://Quà thu hoạch Thám hiểm
			if val, ok := progress[step]; ok && val == 0 {
				gifts := mission.GetStepGift(step, user)

				var wg sync.WaitGroup
				wg.Add(2)
				progress[step] = 1
				uNewbie.StepGift = util.JsonEndCode(progress)
				go func() {
					c.DB.Save(&uNewbie)
					wg.Done()
				}()
				go user.UpdateGifts(gifts, logtype.GIFT_NEWBIE, 0, &wg)
				wg.Wait()

				c.DataResponse = iris.Map{"code": 1, "data": iris.Map{constants.GIFT: gifts}}
			} else {
				c.DataResponse = iris.Map{"code": -1, "msg": "Gift invalid"}
			}

		case 4://Quay tinh cầu đỏ
			if val, ok := progress[step]; ok && val == 0 {
				gifts := mission.GetStepGift(step, user)


				uSum := models.HfUserSummon{}
				uSum.Get(3, user)

				var wg sync.WaitGroup
				wg.Add(3)
				progress[step] = 1
				uNewbie.StepGift = util.JsonEndCode(progress)
				go func() {
					//c.DB.Save(&uNewbie) //Todo cho chạy cập nhật song song luôn vị đụng luồng
					uNewbie.Complete(2, user, 1)
					uNewbie.CheckBigGiftDaily(user)

					wg.Done()
				}()
				go uSum.SetQuantity(-1, 0, logtype.GIFT_NEWBIE, 0, user, &wg)
				go user.UpdateGifts(gifts, logtype.GIFT_NEWBIE, 0, &wg)
				wg.Wait()

				c.DataResponse = iris.Map{"code": 1, "data": iris.Map{constants.GIFT: []interface{}{gifts}}}
			} else {
				c.DataResponse = iris.Map{"code": -1, "msg": "Gift invalid"}
			}

		case 5, 6://Quay vòng quay tím
			if val, ok := progress[step]; ok && val == 0 {
				gifts := mission.GetStepGift(step, user)


				uSum := models.HfUserSummon{}
				uSum.Get(2, user)

				var wg sync.WaitGroup
				wg.Add(3)
				progress[step] = 1
				uNewbie.StepGift = util.JsonEndCode(progress)
				go func() {
					c.DB.Save(&uNewbie)
					wg.Done()
				}()
				go uSum.SetQuantity(-1, 0, logtype.GIFT_NEWBIE, 0, user, &wg)
				go user.UpdateGifts(gifts, logtype.GIFT_NEWBIE, 0, &wg)
				wg.Wait()

				c.DataResponse = iris.Map{"code": 1, "data": iris.Map{constants.GIFT: []interface{}{gifts}}}
			} else {
				c.DataResponse = iris.Map{"code": -1, "msg": "Gift invalid"}
			}


		default:
			c.DataResponse = iris.Map{"code": -1, "msg": "Step invalid"}
		}
	}
}



// /newbie/mission/info
func (c *NewbieController) GetMissionInfo() {
	user := c.User

	uMission := models.HfUserNewbie{}
	uMission.Get(user)

	mission := models.HfMissionNewbie{}
	missionDaily, _, _ , _:= mission.GetProgress(user)
	bigGift := mission.GetBigGift(user)

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

	timeReset := 99999

	c.DataResponse = iris.Map{"code": 1, "data": iris.Map{
		"missions":   dataMission,
		"status":     util.StatusGift(uMission.BigGift),
		"big_gift":   bigGift,
		"time_reset": timeReset,
	}}
}

// /newbie/mission/complete
func (c *NewbieController) PostMissionComplete(form formValue, ctx iris.Context) {
	if c.validToken(ctx) {
		user := c.User
		missionId := util.ParseInt(form("mission_id"))

		uMission := models.HfUserNewbie{}
		uMission.Get(user)

		uMission.Complete(missionId, user, 1)

		uMission.CheckBigGiftDaily(user)

		c.DataResponse = iris.Map{"code": 1, "msg": "Success"}
	}
}

// /newbie/mission/gift
func (c *NewbieController) PostMissionGift(form formValue, ctx iris.Context) {
	if c.validToken(ctx) {
		user := c.User
		missionId := util.ParseInt(form("mission_id"))

		uMission := models.HfUserNewbie{}
		uMission.Get(user)

		progressGift := map[int]int{}
		util.JsonDecodeObject(uMission.Gift, &progressGift)

		if status, ok := progressGift[missionId]; !ok || status != 0 {
			c.DataResponse = iris.Map{"code": -1, "msg": "Mission or gift invalid"}
			return
		}

		mission := models.HfMissionNewbie{}
		gifts := mission.GetMissionGift(missionId, user)

		//update status, gift
		progressGift[missionId] = 1
		uMission.Gift = util.JsonEndCode(progressGift)

		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			c.DB.Save(&uMission)
			wg.Done()
		}()
		go user.UpdateGifts(gifts, logtype.GIFT_MISSION_NEWBIE, 0, &wg)
		wg.Wait()

		c.DataResponse = iris.Map{"code": 1, "msg": "Success", "data": iris.Map{constants.GIFT: gifts}}
	}
}

// /newbie/mission/biggift
func (c *NewbieController) PostMissionBiggift(form formValue, ctx iris.Context) {
	if c.validToken(ctx) {
		user := c.User

		uMission := models.HfUserNewbie{}
		uMission.Get(user)

		progressGift := map[int]int{}
		util.JsonDecodeObject(uMission.Gift, &progressGift)

		if uMission.BigGift != 0 {
			c.DataResponse = iris.Map{"code": -1, "msg": "Big gift invalid"}
			return
		}

		mission := models.HfMissionNewbie{}
		gifts := mission.GetBigGift(user)

		//update status, gift
		uMission.BigGift = 1

		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			c.DB.Save(&uMission)
			wg.Done()
		}()
		go user.UpdateGifts(gifts, logtype.GIFT_BIG_MISSION_NEWBIE, 0, &wg)

		//update cho chắc cú
		wg.Add(1)
		go user.SetNewbieStep(-1, &wg)
		wg.Wait()

		c.DataResponse = iris.Map{"code": 1, "msg": "Success", "data": iris.Map{constants.GIFT: gifts}}
	}
}
