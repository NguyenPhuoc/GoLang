package models

type HfPetFate struct {
	PetId     uint `gorm:"primary_key"`
	PetFate   uint `gorm:"primary_key"`
	BonusBy   string
	BonusType string
	Bonus     string
}

func (HfPetFate) TableName() string {
	return "hf_pet_fate"
}
