package models

type HfUserCampaignNode struct {
	UserId  string `gorm:"primary_key"`
	Node    uint   `gorm:"default:1"`
	MapId   uint   `gorm:"default:1"`
	MapDiff uint   `gorm:"default:1"`
}

func (HfUserCampaignNode) TableName() string {
	return "hf_user_campaign_node"
}

func (ucn *HfUserCampaignNode) Get(u HfUser) HfUserCampaignNode {
	ucn.UserId = u.UserId

	count := 0
	u.DB.Where(ucn).Find(&ucn).Count(&count)

	if count == 0 {
		u.DB.Save(&ucn)
	}

	return *ucn
}
