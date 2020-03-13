package models

type HfUserCampaignLog struct {
	UserId  string
	Node    uint
	MapId   uint
	MapDiff uint
	Win     uint8
	Data    string
}

func (HfUserCampaignLog) TableName() string {
	return "hf_user_campaign_log"
}
