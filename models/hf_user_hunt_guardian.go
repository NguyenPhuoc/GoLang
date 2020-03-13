package models

import (
	"github.com/kataras/iris"
)

type HfUserHuntGuardian struct {
	UserId        string `gorm:"primary_key"`
	GuardianId    uint16 `gorm:"primary_key"`
	GuardianLevel uint16 `gorm:"default:1"`
	GuardianHp    int
}

func (HfUserHuntGuardian) TableName() string {
	return "hf_user_hunt_guardian"
}

func (HfUserHuntGuardian) GetMaps(u HfUser) []iris.Map {
	gh := HfGuardianHunt{}
	guaHunts := gh.GetAll(u)

	data := []iris.Map{}
	for _, gh := range guaHunts {
		if gh.Level == 1 {
			uhg := HfUserHuntGuardian{}
			uhg.Get(gh.GuardianId, u)

			gh = HfGuardianHunt{}
			gh.Find(uhg.GuardianId, uhg.GuardianLevel, u)

			mapConf := gh.GetMap()
			mapConf["id"] = uhg.GuardianId
			mapConf["hp"] = uhg.GuardianHp
			mapConf["max_hp"] = gh.Hp
			delete(mapConf, "guardian_id")

			data = append(data, mapConf)
		}
	}

	return data
}

func (uhg *HfUserHuntGuardian) Get(guardianId uint16, u HfUser) HfUserHuntGuardian {
	uhg.UserId = u.UserId
	uhg.GuardianId = guardianId

	count := 0
	u.DB.Where(uhg).First(&uhg).Count(&count)
	if count == 0 {
		gh := HfGuardianHunt{}
		gh, check := gh.Find(guardianId, 1, u)
		if check {
			uhg.GuardianHp = gh.Hp
			u.DB.Save(&uhg)
		}
	}

	return *uhg
}
