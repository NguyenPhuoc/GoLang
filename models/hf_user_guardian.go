package models

import (
	"GoLang/libraries/constants"
	"GoLang/libraries/util"
	"github.com/kataras/iris"
	"sync"
)

type HfUserGuardian struct {
	Id            string `gorm:"primary_key"`
	UserId        string
	Piece         uint
	Stat          uint8
	GuardianId    uint16
	Level         uint16
	Evolve        uint16
	LevelPassive1 uint16 `gorm:"column:level_passive_1"`
	LevelPassive2 uint16 `gorm:"column:level_passive_2"`
	LevelPassive3 uint16 `gorm:"column:level_passive_3"`
	LevelPassive4 uint16 `gorm:"column:level_passive_4"`
}

func (HfUserGuardian) TableName() string {
	return "hf_user_guardian"
}

func (ug *HfUserGuardian) GetMap() iris.Map {
	return iris.Map{
		"guardian_id":     ug.GuardianId,
		"piece":           ug.Piece,
		"stat":            ug.Stat,
		"level":           ug.Level,
		"evolve":          ug.Evolve,
		"level_passive_1": ug.LevelPassive1,
		"level_passive_2": ug.LevelPassive2,
		"level_passive_3": ug.LevelPassive3,
		"level_passive_4": ug.LevelPassive4,
	}
}

func (ug *HfUserGuardian) GetMaps(u HfUser) []iris.Map {
	ug.UserId = u.UserId

	ugs := []HfUserGuardian{}
	u.DB.Where(ug).Find(&ugs)

	data := []iris.Map{}
	for _, gua := range ugs {
		data = append(data, gua.GetMap())
	}

	return data
}

func (ug *HfUserGuardian) Find(guardianId uint16, u HfUser) (HfUserGuardian, bool) {
	ug.UserId = u.UserId
	ug.GuardianId = guardianId

	count := 0
	if guardianId != 0 {
		u.DB.Where(ug).First(&ug).Count(&count)
	}

	if count == 0 {
		ug.Id = util.UUID()
	}

	return *ug, count != 0
}

func (ug *HfUserGuardian) SetPiece(piece int, kindId uint, typeLog, eventId int, u HfUser, wg *sync.WaitGroup) {
	defer wg.Done()

	ug.Piece = util.QuantityUint(ug.Piece, piece)

	wg.Add(2)
	go u.SaveLog(typeLog, constants.PIECE_GUARDIAN, uint(ug.GuardianId), kindId, eventId, piece, uint64(ug.Piece), "", wg)

	go func() {
		u.DB.Save(&ug)
		wg.Done()
	}()
}


