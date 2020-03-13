package models

import (
	"GoLang/libraries/util"
)

type HfUserNewbie struct {
	UserId   string `gorm:"primary_key"`
	Step     int
	StepGift string
	Progress string
	Gift     string
	BigGift  uint8
}

func (HfUserNewbie) TableName() string {
	return "hf_user_newbie"
}

func (un *HfUserNewbie) Get(u HfUser) HfUserNewbie {
	un.UserId = u.UserId

	count := 0
	u.DB.First(&un).Count(&count)

	if count == 0 {
		un.Step = u.NewbieStep

		mission := HfMissionNewbie{}
		_, progressMission, progressGift, progressStep := mission.GetProgress(u)

		un.Progress = util.JsonEndCode(progressMission)
		un.Gift = util.JsonEndCode(progressGift)
		un.StepGift = util.JsonEndCode(progressStep)
		un.BigGift = 2

		un.Complete(1, u, 1)

		u.DB.Save(&un)
	}

	un.CheckBigGiftDaily(u)

	return *un
}

func (un *HfUserNewbie) CheckBigGiftDaily(u HfUser) {

	if un.BigGift == 2 {
		progressMission := map[int][]int{}
		util.JsonDecodeObject(un.Progress, &progressMission)

		check := true
		for _, progress := range progressMission {
			if progress[0] < progress[1] {

				check = false
				break
			}
		}
		if check {
			un.BigGift = 0
			u.DB.Save(&un)
		}
	}
}

func (un *HfUserNewbie) Complete(missionId int, u HfUser, quantity int) {

	progressMission := map[int][]int{}
	progressGift := map[int]int{}

	util.JsonDecodeObject(un.Progress, &progressMission)
	util.JsonDecodeObject(un.Gift, &progressGift)

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

		un.Progress = util.JsonEndCode(progressMission)
		un.Gift = util.JsonEndCode(progressGift)

		u.DB.Save(&un)
	}
}
