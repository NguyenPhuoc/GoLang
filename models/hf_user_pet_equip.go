package models

import "github.com/jinzhu/gorm"

type HfUserPetEquip struct {
	Id          string `gorm:"primary_key"`
	UserPetId   string
	UserEquipId string
	EquipType   uint8
}

func (HfUserPetEquip) TableName() string {
	return "hf_user_pet_equip"
}

func (upe *HfUserPetEquip) Find(uPetId string, equipType uint8, db *gorm.DB) (HfUserPetEquip, bool) {
	upe.UserPetId = uPetId
	upe.EquipType = equipType

	count := 0
	db.Where(upe).First(&upe).Count(&count)

	return *upe, count != 0
}
