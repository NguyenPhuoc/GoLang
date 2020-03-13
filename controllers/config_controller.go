package controllers

import (
	"GoLang/libraries/util"
	"GoLang/models"
	"fmt"
	"github.com/kataras/iris"
)

type ConfigController struct {
	MyController
}

// /config/all/cache
func (c *ConfigController) GetAllCache() {
	user := c.User

	cam := models.HfCampaign{}
	cam.CacheAll(user)

	camInfo := models.HfCampaignInfo{}
	camInfo.CacheAll(user)

	pet := models.HfPet{}
	pet.CacheAll(user)

	skill := models.HfSkill{}
	skill.CacheAll(user)

	equip := models.HfEquip{}
	equip.CacheAll(user)

	equipStar := models.HfEquipStar{}
	equipStar.CacheAll(user)

	level := models.HfLevel{}
	level.CacheAll(user)

	explore := models.HfExplore{}
	explore.CacheAll(user)

	petLevel := models.HfPetLevel{}
	petLevel.CacheAll(user)

	petStar := models.HfPetStar{}
	petStar.CacheAll(user)

	summon := models.HfSummon{}
	summon.CacheAll(user)

	mission := models.HfMission{}
	mission.CacheAll(user)

	petEvo := models.HfPetEvolve{}
	petEvo.CacheAll(user)

	petSkill := models.HfPetSkill{}
	petSkill.CacheAll(user)

	rollup := models.HfRollup{}
	rollup.CacheAll(user)

	missionNewbie := models.HfMissionNewbie{}
	missionNewbie.CacheAll(user)

	event := models.HfEvent{}
	event.CacheAll(user)

	equipBonus := models.HfEquipBonus{}
	equipBonus.CacheAll(user)

	lineUpTeam := models.HfConfigLineUpTeam{}
	lineUpTeam.CacheAll(user)

	guardian := models.HfGuardian{}
	guardian.CacheAll(user)

	guardianHunt := models.HfGuardianHunt{}
	guardianHunt.CacheAll(user)

	conf := models.HfConfig{}
	conf.CacheAll(user)

	shop := models.HfShop{}
	shop.CacheAll(user)

	guaAward := models.HfGuardianAward{}
	guaAward.CacheAll(user)

	guaUpgrade := models.HfGuardianUpgrade{}
	guaUpgrade.CacheAll(user)

	giftCode := models.HfGiftCode{}
	giftCode.CacheAll(user)

	whiteList := models.HfWhitelist{}
	whiteList.CacheAll(user)

	petRune := models.HfPetRune{}
	petRune.CacheAll(user)

	arenaShop := models.HfArenaShop{}
	arenaShop.CacheAll(user)

	arenaAward := models.HfArenaAward{}
	arenaAward.CacheAll(user)

	arenaPvpAward := models.HfArenaPvpAward{}
	arenaPvpAward.CacheAll(user)

	server := models.HfServer{}
	server.CacheAll(user)


	c.DataResponse = iris.Map{"code": 1, "msg": "Done!!!"}
	/*panic: runtime error: invalid memory address or nil pointer dereference
	[signal SIGSEGV: segmentation violation code=0x1 addr=0x2a8 pc=0xc069a7]

	goroutine 101825 [running]:
	go.mongodb.org/mongo-driver/mongo.newDatabase(0x0, 0xf5ef40, 0x3, 0x0, 0x0, 0x0, 0xc000660f50)
			/data/www/pokiwarh5/src/go.mongodb.org/mongo-driver/mongo/database.go:46 +0x67
	go.mongodb.org/mongo-driver/mongo.(*Client).Database(0x0, 0xf5ef40, 0x3, 0x0, 0x0, 0x0, 0xc000294640)
			/data/www/pokiwarh5/src/go.mongodb.org/mongo-driver/mongo/client.go:648 +0x5d
	GoLang/models.(*HfUser).SaveLogMongo(0xc000153400, 0xf6ee23, 0x12, 0xe5fbe0, 0xc000cc37a0, 0xc000a0f9b0)
			/data/www/pokiwarh5/src/GoLang/models/hf_user.go:518 +0x89
	created by GoLang/models.(*HfUser).SaveLog.func1
			/data/www/pokiwarh5/src/GoLang/models/hf_user.go:511 +0x6ad
	exit status 2
	*/
}

// /config/gen/giftcode
func (c *ConfigController) GetGenGiftcode(ctx iris.Context) {
	c.IsMetaData = true
	user := c.User
	gc := models.HfGiftCode{}
	gcs := gc.GetAll(user)

	for _, g := range gcs {

		if g.Quantity > 1 {

			count := 0
			c.DB.Model(&models.HfGiftCodeItems{}).Where("code = ?", g.Code).Count(&count)
			if count < g.Quantity {
				quan := g.Quantity - count

				codes := []string{}
				fmt.Println(quan, util.JsonEndCode(g))

				for {
					code := util.RanCode(6)
					if !util.InArray(code, codes) {
						codes = append(codes, code)
						//fmt.Println(code)

						inserts := models.HfGiftCodeItems{Code: g.Code, GiftCode: code}
						go func() {
							c.DB.Create(&inserts)
						}()

					}

					if len(codes) >= quan {
						break
					}
				}

				fmt.Println(util.JsonEndCode(g))
			}
		}
	}

	c.DataResponse = iris.Map{"code": 1, "msg": "success"}
}

// /config/campaign
func (c *ConfigController) GetCampaign(ctx iris.Context) {
	c.IsMetaData = true

	//Todo campaign
	cam := models.HfCampaign{}
	campaigns := cam.GetAll(c.User)

	campaignReturn := []iris.Map{}
	for _, val := range campaigns {

		campaignReturn = append(campaignReturn, val.GetMap())
	}

	//Todo campaignInfo
	camInfo := models.HfCampaignInfo{}
	campaignInfos := camInfo.GetAll(c.User)

	campaignInfoReturn := []iris.Map{}
	for _, val := range campaignInfos {

		campaignInfoReturn = append(campaignInfoReturn, val.GetMap())
	}

	c.DataResponse = iris.Map{"code": 1, "data": iris.Map{
		"campaign":      campaignReturn,
		"campaign_info": campaignInfoReturn,
	}}
}

// /config/campaign/v2
func (c *ConfigController) GetCampaignV2(ctx iris.Context) {
	c.IsMetaData = true

	//Todo campaign
	cam := models.HfCampaign{}
	campaigns := cam.GetAll(c.User)

	campaignReturn := []iris.Map{}
	for _, val := range campaigns {

		campaignReturn = append(campaignReturn, val.GetMap())
	}

	//Todo campaignInfo
	camInfo := models.HfCampaignInfo{}
	campaignInfos := camInfo.GetAll(c.User)

	campaignInfoReturn := []iris.Map{}
	for _, val := range campaignInfos {

		campaignInfoReturn = append(campaignInfoReturn, val.GetMap())
	}

	c.IsEncrypt = true
	c.DataResponse = iris.Map{"code": 1, "data": iris.Map{
		"campaign":      campaignReturn,
		"campaign_info": campaignInfoReturn,
	}}
}

// /config/pet
func (c *ConfigController) GetPet(ctx iris.Context) {
	c.IsMetaData = true

	pet := models.HfPet{}
	pets := pet.GetAll(c.User)

	petReturn := []iris.Map{}
	for _, val := range pets {

		petReturn = append(petReturn, val.GetMap())
	}

	c.DataResponse = iris.Map{"code": 1, "data": petReturn}
}

// /config/pet/v2
func (c *ConfigController) GetPetV2(ctx iris.Context) {
	c.IsMetaData = true

	pet := models.HfPet{}
	pets := pet.GetAll(c.User)

	petReturn := []iris.Map{}
	for _, val := range pets {

		petReturn = append(petReturn, val.GetMap())
	}

	c.IsEncrypt = true
	c.DataResponse = iris.Map{"code": 1, "data": petReturn}
}

// /config/pet/star
func (c *ConfigController) GetPetStar(ctx iris.Context) {
	c.IsMetaData = true

	petStar := models.HfPetStar{}
	petStars := petStar.GetAll(c.User)

	petStarReturn := []iris.Map{}
	for _, val := range petStars {

		petStarReturn = append(petStarReturn, val.GetMap())
	}

	c.DataResponse = iris.Map{"code": 1, "data": petStarReturn}
}

// /config/skill
func (c *ConfigController) GetSkill(ctx iris.Context) {
	c.IsMetaData = true

	skill := models.HfSkill{}
	skills := skill.GetAll(c.User)

	skillReturn := []iris.Map{}
	for _, val := range skills {

		skillReturn = append(skillReturn, val.GetMap())
	}

	c.DataResponse = iris.Map{"code": 1, "data": skillReturn}
}

// /config/skill/v2
func (c *ConfigController) GetSkillV2(ctx iris.Context) {
	c.IsMetaData = true

	skill := models.HfSkill{}
	skills := skill.GetAll(c.User)

	skillReturn := []iris.Map{}
	for _, val := range skills {

		skillReturn = append(skillReturn, val.GetMap())
	}

	c.IsEncrypt = true
	c.DataResponse = iris.Map{"code": 1, "data": skillReturn}
}

// /config/equip
func (c *ConfigController) GetEquip(ctx iris.Context) {
	c.IsMetaData = true

	e := models.HfEquip{}
	equips := e.GetAll(c.User)

	s := models.HfEquipStar{}
	starts := s.GetAll(c.User)

	equipReturn := []iris.Map{}
	for _, e := range equips {
		for _, s := range starts {
			if e.Quality == s.Quality && e.Star == s.Star {
				obj := e.GetMap()
				obj["gold_upgrade"] = s.Gold
				obj["sell"] = s.Sell

				equipReturn = append(equipReturn, obj)
				break
			}
		}
	}

	equipBonus := models.HfEquipBonus{}
	ebs := equipBonus.GetAll(c.User)

	equipBonusReturn := []iris.Map{}
	for _, eb := range ebs {

		equipBonusReturn = append(equipBonusReturn, eb.GetMap())
	}

	c.DataResponse = iris.Map{"code": 1, "data": equipReturn, "data_bonus": equipBonusReturn}
}

// /config/equip/v2
func (c *ConfigController) GetEquipV2(ctx iris.Context) {
	c.IsMetaData = true

	e := models.HfEquip{}
	equips := e.GetAll(c.User)

	s := models.HfEquipStar{}
	starts := s.GetAll(c.User)

	equipReturn := []iris.Map{}
	for _, e := range equips {
		for _, s := range starts {
			if e.Quality == s.Quality && e.Star == s.Star {
				obj := e.GetMap()
				obj["gold_upgrade"] = s.Gold
				obj["sell"] = s.Sell

				equipReturn = append(equipReturn, obj)
				break
			}
		}
	}

	equipBonus := models.HfEquipBonus{}
	ebs := equipBonus.GetAll(c.User)

	equipBonusReturn := []iris.Map{}
	for _, eb := range ebs {

		equipBonusReturn = append(equipBonusReturn, eb.GetMap())
	}

	c.IsEncrypt = true
	c.DataResponse = iris.Map{"code": 1, "data": equipReturn, "data_bonus": equipBonusReturn}
}

// /config/level
func (c *ConfigController) GetLevel(ctx iris.Context) {
	c.IsMetaData = true

	level := models.HfLevel{}
	levels := level.GetAll(c.User)

	levelsReturn := []iris.Map{}
	for _, val := range levels {

		levelsReturn = append(levelsReturn, val.GetMap())
	}

	c.DataResponse = iris.Map{"code": 1, "data": levelsReturn}
}

// /config/explore
func (c *ConfigController) GetExplore(ctx iris.Context) {
	c.IsMetaData = true

	explore := models.HfExplore{}
	explores := explore.GetAll(c.User)

	exploresReturn := []iris.Map{}
	for _, val := range explores {

		exploresReturn = append(exploresReturn, val.GetMap())
	}
	//Todo có chỗ liên quan HfUserExplore.Get()
	conf := models.HfConfig{}
	stoneRatio, stoneEvoRatio, expRatio, goldRatio := conf.GetExploreRatio(c.User)

	c.DataResponse = iris.Map{"code": 1, "data": iris.Map{
		"explore":         exploresReturn,
		"stone_ratio":     stoneRatio,
		"stone_evo_ratio": stoneEvoRatio,
		"exp_ratio":       expRatio,
		"gold_ratio":      goldRatio,
	}}
}

// /config/line/up/team
func (c *ConfigController) GetLineUpTeam(ctx iris.Context) {
	c.IsMetaData = true

	lineUp := models.HfConfigLineUpTeam{}
	lineUps := lineUp.GetAll(c.User)

	lineUpReturns := []iris.Map{}
	for _, val := range lineUps {

		lineUpReturns = append(lineUpReturns, val.GetMap())
	}

	c.IsEncrypt = true
	c.DataResponse = iris.Map{"code": 1, "data": lineUpReturns}
}

// /config/guardian
func (c *ConfigController) GetGuardian(ctx iris.Context) {
	c.IsMetaData = true

	guardian := models.HfGuardian{}
	guardians := guardian.GetAll(c.User)

	guardianReturns := []iris.Map{}
	for _, val := range guardians {

		guardianReturns = append(guardianReturns, val.GetMap())
	}

	c.IsEncrypt = true
	c.DataResponse = iris.Map{"code": 1, "data": guardianReturns}
}
