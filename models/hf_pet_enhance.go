package models

import "github.com/jinzhu/gorm"

type HfPetEnhance struct {
	Enhance uint16 `gorm:"primary_key"`
	Star    uint16
	Stone   uint
	Gold    uint
}

func (HfPetEnhance) TableName() string {
	return "hf_pet_enhance"
}

func (pe *HfPetEnhance) GetEnhance(enhance uint16, db *gorm.DB) (HfPetEnhance, bool) {
	pe.Enhance = enhance

	count := 0
	if enhance != 0 {
		db.Where(pe).First(&pe).Count(&count)
	}

	return *pe, count != 0
}