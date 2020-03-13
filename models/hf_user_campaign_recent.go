package models

type HfUserCampaignRecent struct {
	UserId string `gorm:"primary_key"`
	Node   uint
	MapId  uint
	Win    uint8
}

func (HfUserCampaignRecent) TableName() string {
	return "hf_user_campaign_recent"
}

func (ucr *HfUserCampaignRecent) Get(u HfUser) HfUserCampaignRecent {
	ucr.UserId = u.UserId

	u.DB.Where(ucr).First(&ucr)

	return *ucr
}
