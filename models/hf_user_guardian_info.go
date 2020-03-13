package models

import (
	"GoLang/libraries/constants"
	"GoLang/libraries/util"
	"database/sql"
	"sync"
)

type HfUserGuardianInfo struct {
	UserId     string `gorm:"primary_key"`
	Ticket     uint
	Stone      uint
	Flower     uint
	Fruit      uint
	Level      int
	PowerPoint int
	LastLineUp sql.NullString
}

func (HfUserGuardianInfo) TableName() string {
	return "hf_user_guardian_info"
}

func (ugi *HfUserGuardianInfo) Get(u HfUser) (HfUserGuardianInfo, bool) {
	ugi.UserId = u.UserId

	count := 0
	u.DB.First(&ugi).Count(&count)
	if count == 0 {
		u.DB.Save(&ugi)
	}

	return *ugi, count != 0
}

func (ugi *HfUserGuardianInfo) SetTicket(quantity int, itemId interface{}, kindId interface{}, typeLog, eventId int, u HfUser, wg *sync.WaitGroup) {
	defer wg.Done()

	ugi.Ticket = util.QuantityUint(ugi.Ticket, quantity)

	wg.Add(2)
	go u.SaveLog(typeLog, constants.TICKET_GUARDIAN, itemId, kindId, eventId, quantity, uint64(ugi.Ticket), "", wg)

	go func() {
		u.DB.Model(&ugi).Update("ticket", ugi.Ticket)
		wg.Done()
	}()
}

func (ugi *HfUserGuardianInfo) SetStone(quantity int, itemId interface{}, kindId interface{}, typeLog, eventId int, u HfUser, wg *sync.WaitGroup) {
	defer wg.Done()

	ugi.Stone = util.QuantityUint(ugi.Stone, quantity)

	wg.Add(2)
	go u.SaveLog(typeLog, constants.STONE_GUARDIAN, itemId, kindId, eventId, quantity, uint64(ugi.Stone), "", wg)

	go func() {
		u.DB.Model(&ugi).Update("stone", ugi.Stone)
		wg.Done()
	}()
}

func (ugi *HfUserGuardianInfo) SetFlower(quantity int, itemId interface{}, kindId interface{}, typeLog, eventId int, u HfUser, wg *sync.WaitGroup) {
	defer wg.Done()

	ugi.Flower = util.QuantityUint(ugi.Flower, quantity)

	wg.Add(2)
	go u.SaveLog(typeLog, constants.FLOWER_GUARDIAN, itemId, kindId, eventId, quantity, uint64(ugi.Flower), "", wg)

	go func() {
		u.DB.Model(&ugi).Update("flower", ugi.Flower)
		wg.Done()
	}()
}

func (ugi *HfUserGuardianInfo) SetFruit(quantity int, itemId interface{}, kindId interface{}, typeLog, eventId int, u HfUser, wg *sync.WaitGroup) {
	defer wg.Done()

	ugi.Fruit = util.QuantityUint(ugi.Fruit, quantity)

	wg.Add(2)
	go u.SaveLog(typeLog, constants.FRUIT_GUARDIAN, itemId, kindId, eventId, quantity, uint64(ugi.Fruit), "", wg)

	go func() {
		u.DB.Model(&ugi).Update("fruit", ugi.Fruit)
		wg.Done()
	}()
}
