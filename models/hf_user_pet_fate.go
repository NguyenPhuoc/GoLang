package models

type HfUserPetFate struct {
	Id         string `gorm:"primary_key"`
	UserId     string
	PetId      uint
	PetFate    uint
	BonusType  string
	BonusValue uint
}

func (HfUserPetFate) TableName() string {
	return "hf_user_pet_fate"
}
