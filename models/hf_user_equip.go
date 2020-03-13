package models

import (
	"github.com/kataras/iris"
)

type HfUserEquip struct {
	Id       string `gorm:"primary_key"`
	UserId   string
	EquipId  uint16
	Quantity uint16
	Used     uint16
}

func (HfUserEquip) TableName() string {
	return "hf_user_equip"
}

func (ue *HfUserEquip) Find(equipId uint16, u HfUser) (HfUserEquip, bool) {
	ue.UserId = u.UserId
	ue.EquipId = equipId

	count := 0
	if equipId != 0 {
		u.DB.Where(ue).First(&ue).Count(&count)
	}

	return *ue, count != 0
}

func (ue *HfUserEquip) GetMap() iris.Map {
	return iris.Map{
		"equip_id": ue.EquipId,
		"quantity": ue.Quantity,
		"used":     ue.Used,
	}
}

func (ue *HfUserEquip) GetEquipMap(u HfUser) []iris.Map {
	equips := []HfUserEquip{}
	where := HfUserEquip{UserId: u.UserId}
	u.DB.Where(where).Find(&equips)

	uEquip := make([]iris.Map, len(equips))
	for i, val := range equips {
		uEquip[i] = val.GetMap()
	}

	return uEquip
}
