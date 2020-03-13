package controllers

import (
	"GoLang/libraries/constants"
	"GoLang/libraries/logtype"
	"GoLang/libraries/util"
	"GoLang/models"
	"fmt"
	"github.com/kataras/iris"
	"math"
	"sync"
)

type PetController struct {
	MyController
}

// /pet/upgrade/info
func (c *PetController) GetUpgradeInfo(ctx iris.Context) {
	uStones := c.User.GetStones()
	c.DataResponse = iris.Map{
		"code": 1,
		"data": iris.Map{
			"stones": uStones,
		},
	}
}

// /pet/upgrade
func (c *PetController) PostUpgrade(form formValue, ctx iris.Context) {
	if c.validToken(ctx) {
		user := c.User
		petId := uint16(util.ParseInt(form("pet_id")))
		upType := form("type") //kiểu nâng cấp

		code, msg := -1, ``
		data := iris.Map{}

		uPet, checkUPet := user.GetPet(petId)

		pet := models.HfPet{}
		pet, checkPet := pet.Find(petId, user)

		var wg sync.WaitGroup

		//Todo Kiểm tra pet trong thư viện để nâng cấp
		if !checkPet {
			c.DataResponse = iris.Map{"code": -1, "msg": "Pet Invalid"}
			return
		}

		switch upType {
		case "level":
			max_level := util.ParseInt(form("max_level"))

			uStones := user.GetStones() //đá của user
			petLevel := models.HfPetLevel{}

			currentLevel := uPet.Level
			nextLevel := currentLevel + 1

			levelConf, checkLvCf := petLevel.GetLevel(nextLevel, user)

			checkMaxLevel := petLevel.GetMaxLevel()[pet.Type] >= nextLevel

			//Todo kiểm tra level max, levelConf, đá theo hệ(type), gold, cấp tiến hóa
			if checkUPet && checkMaxLevel && checkLvCf && uStones[pet.Branch] >= levelConf.Stone && user.Gold >= levelConf.Gold && uPet.Evolve >= levelConf.Evolve {
				if max_level == 1 {
					allLevels := petLevel.GetAll(user)
					userGold := user.Gold
					userStone := uStones[pet.Branch]

					for _, lv := range allLevels {
						if lv.Level > uPet.Level && lv.Evolve <= uPet.Evolve && userStone >= lv.Stone && userGold >= lv.Gold {
							uPet.Level = lv.Level

							userGold -= lv.Gold
							userStone -= lv.Stone
						}
					}

					c.DB.Save(&uPet)
					//trừ gold
					goldFee := int(user.Gold-userGold) * -1
					wg.Add(2)
					go user.SetGold(goldFee, uint(uPet.PetId), uint(nextLevel), logtype.UPGRADE_MAX_LEVEL_PET, 0, &wg)
					//trừ đá
					stoneFee := int(uStones[pet.Branch]-userStone) * -1
					go user.SetStones(pet.Branch, stoneFee, uint(uPet.PetId), logtype.UPGRADE_MAX_LEVEL_PET, 0, &wg)
					wg.Wait()

					//Để trả về hiện thị
					uStones[pet.Branch] = userStone

				} else {
					uPet.Level = nextLevel
					c.DB.Save(&uPet)
					//trừ gold
					goldFee := int(levelConf.Gold) * -1
					wg.Add(2)
					go user.SetGold(goldFee, uint(uPet.PetId), uint(nextLevel), logtype.UPGRADE_LEVEL_PET, 0, &wg)
					//trừ đá
					stoneFee := int(levelConf.Stone) * -1
					go user.SetStones(pet.Branch, stoneFee, uint(uPet.PetId), logtype.UPGRADE_LEVEL_PET, 0, &wg)
					wg.Wait()

					//Để trả về hiện thị
					uStones[pet.Branch] -= levelConf.Stone
				}

				code = 1
				msg = `Success`
				data = iris.Map{"pet": uPet.GetMap(), constants.GOLD: user.Gold, constants.STONES: uStones}
			} else {
				msg = `Not Eligible`
			}
		case "evolve":                  //Todo Tiến hóa
			uStones := user.GetStones() //đá của user
			petEvolve := models.HfPetEvolve{}

			currentEvolve := uPet.Evolve
			nextEvolve := currentEvolve + 1

			evolveConf, checkEvolve := petEvolve.Find(nextEvolve, user)
			//Todo kiểm tra Tiến Hóa max, đá Tiến hóa, gold, level Pet
			if checkUPet && checkEvolve && uStones[constants.EVOLVE] >= evolveConf.Stone && user.Gold >= evolveConf.Gold && uPet.Level >= evolveConf.Level {
				uPet.Evolve = nextEvolve
				c.DB.Save(&uPet)
				//trừ gold
				goldFee := int(evolveConf.Gold) * -1
				wg.Add(2)
				go user.SetGold(goldFee, uint(uPet.PetId), uint(nextEvolve), logtype.UPGRADE_EVOLVE_PET, 0, &wg)
				//trừ đá
				stoneFee := int(evolveConf.Stone) * -1
				go user.SetStones(constants.EVOLVE, stoneFee, uint(uPet.PetId), logtype.UPGRADE_EVOLVE_PET, 0, &wg)
				wg.Wait()

				code = 1
				msg = `Success`
				data = iris.Map{"pet": uPet.GetMap()}
			} else {
				msg = `Not Eligible`
			}
		case "insert": //Todo Ghép Pet
			petInsert := models.HfPetStar{}
			starConf, checkStar := petInsert.GetStar(0, pet.Type, user)

			up := models.HfUserPet{PetId: petId}
			up.UserId = user.UserId
			c.DB.Where(up).First(&up)

			//Todo kiểm tra mảnh pet
			if !checkUPet && checkStar && up.Piece >= starConf.Piece {
				up.Stat = 1
				up.Level = 1
				//trừ mảnh
				pieceFee := int(starConf.Piece) * -1
				wg.Add(1)
				go up.SetPiece(pieceFee, 0, logtype.UPGRADE_INSERT_PET, 0, user, &wg)

				//nhận tài nguyên triệu hồi
				wg.Add(1)
				gifts := util.JsonDecodeMap(starConf.Gift)
				go user.UpdateGifts(gifts, logtype.UPGRADE_INSERT_PET, 0, &wg)
				wg.Wait()

				code = 1
				msg = `Success`
				data = iris.Map{"pet": up.GetPetEquipMap(c.DB), constants.GIFT: gifts}
			} else {
				msg = `Not Eligible`
			}
		case "star":                                    //Todo Đột Phá (Sao)
			uPets := util.JsonDecodeArray(form("pets")) // [3001,3002,3003,3004]
			starConf := models.HfPetStar{}

			currentStar := uPet.Star
			nextStar := currentStar + 1

			starConf, checkStar := starConf.GetStar(nextStar, pet.Type, user)
			//Todo kiểm tra Đột Phá trong petStar.CheckUpgradeStar
			checkUpgradeStar, checkMsg := starConf.CheckUpgradeStar(uPet, pet, uPets, user)
			if checkUPet && checkStar && checkUpgradeStar {
				uPet.Star = nextStar
				//uPet.Enhance = starConf.Enhance //Cấp Thức Tỉnh sẽ được kích hoạt
				//không có save ở đây vì có save ở trong uPet.SetPiece

				//trừ mảnh
				pieceFee := int(starConf.Piece) * -1
				wg.Add(1)
				go uPet.SetPiece(pieceFee, uint(nextStar), logtype.UPGRADE_STAR_PET, 0, user, &wg)
				wg.Wait()

				//Todo Trừ mảnh ở từng pet yêu cầu nếu có
				petsRequest := []iris.Map{}
				petCfg := util.JsonDecodeArray(starConf.Pet)

				petsUserArr := util.ToArrayInt(uPets)
				petsUserArr = util.UniqueInt(petsUserArr)

				//Kiểm tra pet dùng nâng cấp
				for i, petId := range petsUserArr {
					petId := uint16(petId)
					//lấy pet config theo đúng thứ tự
					petRequest := util.InterfaceToMap(petCfg[i])

					//pet cần check
					uPetCheck, _ := user.GetPet(petId)

					//Mảnh
					if val, ok := petRequest["piece"]; ok {
						piecePet := util.ToInt(val)
						pieceFee := piecePet * -1

						wg.Add(1)
						go uPetCheck.SetPiece(pieceFee, uint(nextStar), logtype.MATERIAL_UPGRADED_PET, 0, user, &wg)
						wg.Wait()

						petsRequest = append(petsRequest, uPetCheck.GetPetEquipMap(c.DB))
					}
				}

				code = 1
				msg = `Success`
				data = iris.Map{"pet": uPet.GetPetEquipMap(c.DB), "pet_request": petsRequest}
			} else {
				msg = `Not Eligible ` + checkMsg
			}
		case "skill": //Todo Nâng cấp skill
			skillId := uint16(util.ParseInt(form("skill_id")))
			skillConf := models.HfPetSkill{}

			skills := util.JsonDecodeMap(uPet.Skill)
			currentLevel := uint16(util.ToInt(skills[util.ToString(skillId)]))
			nextLevel := currentLevel + 1

			skillConf, checkSkill := skillConf.Find(skillId, nextLevel, user)
			//Todo kiểm tra Nâng cấp Skill trong skillConf.CheckUpgradeStar
			checkUpgradeStar, checkMsg := skillConf.CheckUpgradeStar(uPet, skillId, user)
			if checkUPet && checkSkill && checkUpgradeStar {
				skills[util.ToString(skillId)] = nextLevel
				uPet.Skill = util.JsonEndCode(skills)
				c.DB.Save(&uPet)

				//Todo Nguyên liệu Gold, Stone
				material := util.JsonDecodeMap(skillConf.Material)
				if val, ok := material[constants.GOLD]; ok {
					goldFee := util.ToInt(val) * -1

					wg.Add(1)
					go user.SetGold(goldFee, skillId, uint(nextLevel), logtype.UPGRADE_SKILL_PET, 0, &wg)
				}

				//Đá tiến hóa
				if val, ok := material[constants.STONES]; ok {
					stone := util.InterfaceToMap(val)
					for key, quan := range stone {
						stoneFee := util.ToInt(quan) * -1
						wg.Add(1)
						go user.SetStones(key, stoneFee, uint(skillId), logtype.UPGRADE_SKILL_PET, 0, &wg)
					}
				}
				wg.Wait()

				code = 1
				msg = `Success`
				data = iris.Map{"pet": uPet.GetPetEquipMap(c.DB)}
			} else {
				msg = `Not Eligible ` + checkMsg
			}
		case "enhance": //Todo Thức Tỉnh ==>> Phần này coment vì thức tỉnh đã được chèn zo lúc đột phá sao ở phía trên
			//petEnhance := models.HfPetEnhance{}
			//
			//currentEnhance := uPet.Enhance
			//nextEnhance := currentEnhance + 1
			//
			//enhanceConf, checkEnhance := petEnhance.GetEnhance(nextEnhance, c.DB)
			////Todo kiểm tra Thức Tỉnh max, Gold, đá Thức Tỉnh, Sao yêu cầu
			//if checkUPet && checkEnhance && user.Gold >= enhanceConf.Gold && uStones[util.THUC_TINH] >= enhanceConf.Stone && uPet.Skill >= enhanceConf.Skill {
			//	uPet.Enhance = nextEnhance
			//	c.DB.Save(&uPet)
			//	//trừ gold
			//	goldFee := int(enhanceConf.Gold) * -1
			//	wg.Add(2)
			//	go user.SetGold(goldFee, uint(uPet.PetId), uint(nextEnhance), logtype.UPGRADE_ENHANCE_PET, &wg)
			//	//trừ đá
			//	stoneFee := int(enhanceConf.Stone) * -1
			//	go user.SetStones(util.THUC_TINH, stoneFee, uint(uPet.PetId), logtype.UPGRADE_ENHANCE_PET, &wg)
			//	wg.Wait()
			//
			//	code = 1
			//	msg = `Success`
			//	data = iris.Map{"pet": uPet.GetMap()}
			//} else {
			//	msg = `Not Eligible`
			//}
		default:
			msg = `Type Invalid!`
		}

		c.DataResponse = iris.Map{"code": code, "msg": msg, "data": data}
	}
}

// /pet/upgrade/v2
func (c *PetController) PostUpgradeV2(form formValue, ctx iris.Context) {
	if c.validToken(ctx) {
		user := c.User
		petId := uint16(util.ParseInt(form("pet_id")))
		upType := form("type") //kiểu nâng cấp
		c.IsEncrypt = true

		code, msg := -1, ``
		data := iris.Map{}

		uPet, checkUPet := user.GetPet(petId)

		pet := models.HfPet{}
		pet, checkPet := pet.Find(petId, user)

		var wg sync.WaitGroup

		//Todo Kiểm tra pet trong thư viện để nâng cấp
		if !checkPet {
			c.DataResponse = iris.Map{"code": -1, "msg": "Pet Invalid"}
			return
		}

		switch upType {
		case "level":
			max_level := util.ParseInt(form("max_level"))

			uStones := user.GetStones() //đá của user
			petLevel := models.HfPetLevel{}

			currentLevel := uPet.Level
			nextLevel := currentLevel + 1

			levelConf, checkLvCf := petLevel.GetLevel(nextLevel, user)

			checkMaxLevel := petLevel.GetMaxLevel()[pet.Type] >= nextLevel

			//Todo kiểm tra level max, levelConf, đá theo hệ(type), gold, cấp tiến hóa
			if checkUPet && checkMaxLevel && checkLvCf && uStones[pet.Branch] >= levelConf.Stone && user.Gold >= levelConf.Gold && uPet.Evolve >= levelConf.Evolve {
				if max_level == 1 {
					allLevels := petLevel.GetAll(user)
					userGold := user.Gold
					userStone := uStones[pet.Branch]

					for _, lv := range allLevels {
						if lv.Level > uPet.Level && lv.Evolve <= uPet.Evolve && userStone >= lv.Stone && userGold >= lv.Gold {
							uPet.Level = lv.Level

							userGold -= lv.Gold
							userStone -= lv.Stone
						}
					}

					c.DB.Save(&uPet)
					//trừ gold
					goldFee := int(user.Gold-userGold) * -1
					wg.Add(2)
					go user.SetGold(goldFee, uint(uPet.PetId), uint(nextLevel), logtype.UPGRADE_MAX_LEVEL_PET, 0, &wg)
					//trừ đá
					stoneFee := int(uStones[pet.Branch]-userStone) * -1
					go user.SetStones(pet.Branch, stoneFee, uint(uPet.PetId), logtype.UPGRADE_MAX_LEVEL_PET, 0, &wg)
					wg.Wait()

					//Để trả về hiện thị
					uStones[pet.Branch] = userStone

				} else {
					uPet.Level = nextLevel
					c.DB.Save(&uPet)
					//trừ gold
					goldFee := int(levelConf.Gold) * -1
					wg.Add(2)
					go user.SetGold(goldFee, uint(uPet.PetId), uint(nextLevel), logtype.UPGRADE_LEVEL_PET, 0, &wg)
					//trừ đá
					stoneFee := int(levelConf.Stone) * -1
					go user.SetStones(pet.Branch, stoneFee, uint(uPet.PetId), logtype.UPGRADE_LEVEL_PET, 0, &wg)
					wg.Wait()

					//Để trả về hiện thị
					uStones[pet.Branch] -= levelConf.Stone
				}

				code = 1
				msg = `Success`
				data = iris.Map{"pet": uPet.GetMap(), constants.GOLD: user.Gold, constants.STONES: uStones}
			} else {
				msg = `Not Eligible`
			}
		case "evolve":                  //Todo Tiến hóa
			uStones := user.GetStones() //đá của user
			petEvolve := models.HfPetEvolve{}

			currentEvolve := uPet.Evolve
			nextEvolve := currentEvolve + 1

			evolveConf, checkEvolve := petEvolve.Find(nextEvolve, user)
			//Todo kiểm tra Tiến Hóa max, đá Tiến hóa, gold, level Pet
			if checkUPet && checkEvolve && uStones[constants.EVOLVE] >= evolveConf.Stone && user.Gold >= evolveConf.Gold && uPet.Level >= evolveConf.Level {
				uPet.Evolve = nextEvolve
				c.DB.Save(&uPet)
				//trừ gold
				goldFee := int(evolveConf.Gold) * -1
				wg.Add(2)
				go user.SetGold(goldFee, uint(uPet.PetId), uint(nextEvolve), logtype.UPGRADE_EVOLVE_PET, 0, &wg)
				//trừ đá
				stoneFee := int(evolveConf.Stone) * -1
				go user.SetStones(constants.EVOLVE, stoneFee, uint(uPet.PetId), logtype.UPGRADE_EVOLVE_PET, 0, &wg)
				wg.Wait()

				code = 1
				msg = `Success`
				data = iris.Map{"pet": uPet.GetMap()}
			} else {
				msg = `Not Eligible`
			}
		case "insert": //Todo Ghép Pet
			petInsert := models.HfPetStar{}
			starConf, checkStar := petInsert.GetStar(0, pet.Type, user)

			up := models.HfUserPet{PetId: petId}
			up.UserId = user.UserId
			c.DB.Where(up).First(&up)

			//Todo kiểm tra mảnh pet
			if !checkUPet && checkStar && up.Piece >= starConf.Piece {
				up.Stat = 1
				up.Level = 1
				//trừ mảnh
				pieceFee := int(starConf.Piece) * -1
				wg.Add(1)
				go up.SetPiece(pieceFee, 0, logtype.UPGRADE_INSERT_PET, 0, user, &wg)

				//nhận tài nguyên triệu hồi
				wg.Add(1)
				gifts := util.JsonDecodeMap(starConf.Gift)
				go user.UpdateGifts(gifts, logtype.UPGRADE_INSERT_PET, 0, &wg)
				wg.Wait()

				code = 1
				msg = `Success`
				data = iris.Map{"pet": up.GetPetEquipMap(c.DB), constants.GIFT: gifts}
			} else {
				msg = `Not Eligible`
			}
		case "star":                                    //Todo Đột Phá (Sao)
			uPets := util.JsonDecodeArray(form("pets")) // [3001,3002,3003,3004]
			starConf := models.HfPetStar{}

			currentStar := uPet.Star
			nextStar := currentStar + 1

			starConf, checkStar := starConf.GetStar(nextStar, pet.Type, user)
			//Todo kiểm tra Đột Phá trong petStar.CheckUpgradeStar
			checkUpgradeStar, checkMsg := starConf.CheckUpgradeStar(uPet, pet, uPets, user)
			if checkUPet && checkStar && checkUpgradeStar {
				uPet.Star = nextStar
				//uPet.Enhance = starConf.Enhance //Cấp Thức Tỉnh sẽ được kích hoạt
				//không có save ở đây vì có save ở trong uPet.SetPiece

				//trừ mảnh
				pieceFee := int(starConf.Piece) * -1
				gemSoulFee := int(starConf.GemSoul) * -1
				wg.Add(2)
				go uPet.SetPiece(pieceFee, uint(nextStar), logtype.UPGRADE_STAR_PET, 0, user, &wg)
				go uPet.SetGemSoul(gemSoulFee, logtype.UPGRADE_STAR_PET, 0, user, &wg)
				wg.Wait()

				//Todo Trừ mảnh ở từng pet yêu cầu nếu có
				petsRequest := []iris.Map{}
				petCfg := util.JsonDecodeArray(starConf.Pet)

				petsUserArr := util.ToArrayInt(uPets)
				petsUserArr = util.UniqueInt(petsUserArr)

				//Kiểm tra pet dùng nâng cấp
				for i, petId := range petsUserArr {
					petId := uint16(petId)
					//lấy pet config theo đúng thứ tự
					petRequest := util.InterfaceToMap(petCfg[i])

					//pet cần check
					uPetCheck, _ := user.GetPet(petId)

					//Mảnh
					if val, ok := petRequest["piece"]; ok {
						piecePet := util.ToInt(val)
						pieceFee := piecePet * -1

						wg.Add(1)
						go uPetCheck.SetPiece(pieceFee, uint(nextStar), logtype.MATERIAL_UPGRADED_PET, 0, user, &wg)
						wg.Wait()

						petsRequest = append(petsRequest, uPetCheck.GetPetEquipMap(c.DB))
					}
				}

				code = 1
				msg = `Success`
				data = iris.Map{"pet": uPet.GetPetEquipMap(c.DB), "pet_request": petsRequest}
			} else {
				msg = `Not Eligible ` + checkMsg
			}
		case "skill": //Todo Nâng cấp skill
			skillId := uint16(util.ParseInt(form("skill_id")))
			skillConf := models.HfPetSkill{}

			skills := util.JsonDecodeMap(uPet.Skill)
			currentLevel := uint16(util.ToInt(skills[util.ToString(skillId)]))
			nextLevel := currentLevel + 1

			skillConf, checkSkill := skillConf.Find(skillId, nextLevel, user)
			//Todo kiểm tra Nâng cấp Skill trong skillConf.CheckUpgradeStar
			checkUpgradeStar, checkMsg := skillConf.CheckUpgradeStar(uPet, skillId, user)
			if checkUPet && checkSkill && checkUpgradeStar {
				skills[util.ToString(skillId)] = nextLevel
				uPet.Skill = util.JsonEndCode(skills)
				c.DB.Save(&uPet)

				//Todo Nguyên liệu Gold, Stone
				material := util.JsonDecodeMap(skillConf.Material)
				if val, ok := material[constants.GOLD]; ok {
					goldFee := util.ToInt(val) * -1

					wg.Add(1)
					go user.SetGold(goldFee, skillId, uint(nextLevel), logtype.UPGRADE_SKILL_PET, 0, &wg)
				}

				//Đá tiến hóa
				if val, ok := material[constants.STONES]; ok {
					stone := util.InterfaceToMap(val)
					for key, quan := range stone {
						stoneFee := util.ToInt(quan) * -1
						wg.Add(1)
						go user.SetStones(key, stoneFee, uint(skillId), logtype.UPGRADE_SKILL_PET, 0, &wg)
					}
				}
				wg.Wait()

				code = 1
				msg = `Success`
				data = iris.Map{"pet": uPet.GetPetEquipMap(c.DB)}
			} else {
				msg = `Not Eligible ` + checkMsg
			}
		case "enhance": //Todo Thức Tỉnh ==>> Phần này coment vì thức tỉnh đã được chèn zo lúc đột phá sao ở phía trên
			//petEnhance := models.HfPetEnhance{}
			//
			//currentEnhance := uPet.Enhance
			//nextEnhance := currentEnhance + 1
			//
			//enhanceConf, checkEnhance := petEnhance.GetEnhance(nextEnhance, c.DB)
			////Todo kiểm tra Thức Tỉnh max, Gold, đá Thức Tỉnh, Sao yêu cầu
			//if checkUPet && checkEnhance && user.Gold >= enhanceConf.Gold && uStones[util.THUC_TINH] >= enhanceConf.Stone && uPet.Skill >= enhanceConf.Skill {
			//	uPet.Enhance = nextEnhance
			//	c.DB.Save(&uPet)
			//	//trừ gold
			//	goldFee := int(enhanceConf.Gold) * -1
			//	wg.Add(2)
			//	go user.SetGold(goldFee, uint(uPet.PetId), uint(nextEnhance), logtype.UPGRADE_ENHANCE_PET, &wg)
			//	//trừ đá
			//	stoneFee := int(enhanceConf.Stone) * -1
			//	go user.SetStones(util.THUC_TINH, stoneFee, uint(uPet.PetId), logtype.UPGRADE_ENHANCE_PET, &wg)
			//	wg.Wait()
			//
			//	code = 1
			//	msg = `Success`
			//	data = iris.Map{"pet": uPet.GetMap()}
			//} else {
			//	msg = `Not Eligible`
			//}
		default:
			msg = `Type Invalid!`
		}

		c.DataResponse = iris.Map{"code": code, "msg": msg, "data": data}
	}
}

// /pet/equip
func (c *PetController) GetEquip(ctx iris.Context) {
	ue := models.HfUserEquip{}
	uEquip := ue.GetEquipMap(c.User)

	c.DataResponse = iris.Map{
		"code": 1,
		"data": iris.Map{
			"equip": uEquip,
		},
	}
}

// /pet/wear
func (c *PetController) PostWear(form formValue, ctx iris.Context) {
	if c.validToken(ctx) {
		user := c.User

		code, msg, data := -1, ``, iris.Map{}

		petId := uint16(util.ParseInt(form("pet_id")))
		equipIds := util.JsonDecodeArray(form("equip_id"))

		for _, val := range equipIds {

			equipId := uint16(val.(float64))

			var wg sync.WaitGroup

			//Todo Lấy thông tin ban đầu
			wg.Add(3)
			//Trang bị của user
			uEquip, checUe := models.HfUserEquip{}, false
			go func() {
				uEquip, checUe = uEquip.Find(equipId, user)
				wg.Done()
			}()

			//Trang bị config
			eConf, checkEConf := models.HfEquip{}, false
			go func() {
				eConf, checkEConf = eConf.Find(equipId, user)
				wg.Done()
			}()

			//Pet của user
			uPet, checUp := models.HfUserPet{}, false
			go func() {
				uPet, checUp = uPet.Find(petId, user)
				wg.Done()
			}()
			wg.Wait()

			//Điều kiện cơ bản để bắt đầu có thể mặc đồ
			check := false
			if checUe && checkEConf && checUp && uEquip.Used < uEquip.Quantity {
				check = true
			}

			//Todo kiểm tra đã mặc trang bị chưa và mặc
			if check {
				upe, checkUpe := models.HfUserPetEquip{}, false
				upe, checkUpe = upe.Find(uPet.Id, eConf.Type, c.DB)
				if !checkUpe {
					upe.Id = util.UUID()
				}
				upe.UserEquipId = uEquip.Id
				c.DB.Save(&upe)
			}
		}
		code, msg = 1, `Success`

		c.DataResponse = iris.Map{
			"code": code,
			"msg":  msg,
			"data": data,
		}
	}
}

// /pet/sell/equip
func (c *PetController) PostSellEquip(form formValue, ctx iris.Context) {
	if c.validToken(ctx) {
		user := c.User

		code, msg, data := -1, ``, iris.Map{}

		equipId := uint16(util.ParseInt(form("equip_id")))
		quantity := uint16(util.ParseInt(form("quantity")))
		if quantity < 1 {
			quantity = 1
		}
		var wg sync.WaitGroup

		//Todo Lấy thông tin ban đầu
		wg.Add(2)
		//Trang bị của user
		uEquip, checUe := models.HfUserEquip{}, false
		go func() {
			uEquip, checUe = uEquip.Find(equipId, user)
			wg.Done()
		}()

		//Trang bị config
		eConf, checkEConf := models.HfEquip{}, false
		go func() {
			eConf, checkEConf = eConf.Find(equipId, user)
			wg.Done()
		}()
		wg.Wait()

		//Giá Bán
		eStar := models.HfEquipStar{}
		eStar, checkEStar := eStar.Find(eConf.Quality, eConf.Star, user)

		//Điều kiện cơ bản để có thể bán
		quantityCanSell := uEquip.Quantity - uEquip.Used
		if checUe && checkEConf && checkEStar && quantityCanSell >= quantity {
			code = 1
		}

		//Todo bán trang bị và lưu log
		if code == 1 {
			wg.Add(2)
			//Giảm Số lượng
			go func() {
				uEquip.Quantity -= quantity
				c.DB.Save(&uEquip)
				wg.Done()
			}()

			//Cộng gold
			goldSell := int(eStar.Sell) * int(quantity)
			go user.SetGold(goldSell, uint(equipId), uint(quantity), logtype.SELL_EQUIP_PET, 0, &wg)

			wg.Wait()
			data = iris.Map{
				"gold":      goldSell,
				"user_gold": user.Gold,
			}

			msg = `Success`
		}

		c.DataResponse = iris.Map{
			"code": code,
			"msg":  msg,
			"data": data,
		}
	}
}

// /pet/remove/equip
func (c *PetController) PostRemoveEquip(form formValue, ctx iris.Context) {
	if c.validToken(ctx) {
		user := c.User

		code, msg, data := -1, ``, iris.Map{}

		petId := uint16(util.ParseInt(form("pet_id")))
		typeEquip := uint8(util.ParseInt(form("type")))

		//Lấy pet của user
		uPet := models.HfUserPet{}
		uPet, checkPet := uPet.Find(petId, user)
		if checkPet {
			//xóa trang bị đang mặc của pet
			upe := models.HfUserPetEquip{UserPetId: uPet.Id, EquipType: typeEquip}
			if typeEquip != 0 {
				c.DB.Delete(&upe, "user_pet_id = ? and equip_type = ?", uPet.Id, typeEquip)
			} else {
				c.DB.Delete(&upe, "user_pet_id = ?", uPet.Id)
			}

			code, msg = 1, `Success`
		}

		c.DataResponse = iris.Map{
			"code": code,
			"msg":  msg,
			"data": data,
		}
	}
}

// /pet/upgrade/equip
func (c *PetController) PostUpgradeEquip(form formValue, ctx iris.Context) {
	if c.validToken(ctx) {
		user := c.User

		code, msg, data := -1, ``, iris.Map{}

		equipId := uint16(util.ParseInt(form("equip_id")))
		quantity := uint16(util.ParseInt(form("quantity")))

		var wg sync.WaitGroup
		wg.Add(2)

		equip, checkEquip := models.HfEquip{}, false
		uEquip, checkUequip := models.HfUserEquip{}, false

		go func() {
			equip, checkEquip = equip.Find(equipId, user)
			wg.Done()
		}()

		go func() {
			uEquip, checkUequip = uEquip.Find(equipId, user)
			wg.Done()
		}()

		wg.Wait()

		nextEquip, checkNextEq := models.HfEquip{}, false
		equipStar, checkEStar := models.HfEquipStar{}, false
		if equip.NextId != 0 {
			nextEquip, checkNextEq = nextEquip.Find(equip.NextId, user)

			if checkNextEq {
				equipStar, checkEStar = equipStar.Find(nextEquip.Quality, nextEquip.Star, user)
			}
		}

		//Điều kiện nâng cấp
		quantityAvailable := uEquip.Quantity - uEquip.Used
		goldAvailable := user.Gold

		goldRequest := equipStar.Gold * uint(quantity)
		quantityRequest := equipStar.Quantity * quantity

		checkGold := goldAvailable >= goldRequest
		checkQuantity := quantityAvailable >= quantityRequest

		if quantity > 0 && checkEquip && equip.NextId != 0 && checkUequip && checkEStar && checkGold && checkQuantity {
			//Todo Tiến hành Tăng cấp
			wg.Add(4)

			//Trừ Gold
			goldFee := int(goldRequest) * -1
			go user.SetGold(goldFee, uint(equip.Id), uint(equip.NextId), logtype.UPGRADE_EQUIP_PET, 0, &wg)

			//Trừ Trang Bị
			equipFee := int(quantityRequest) * -1
			go user.SetPetEquip(equip.Id, equipFee, uint(equip.NextId), logtype.UPGRADE_EQUIP_PET, 0, &wg)
			uEquip.Quantity -= quantityRequest //trừ để hiển thị
			//Cộng Trang Bị
			go user.SetPetEquip(equip.NextId, int(quantity), uint(equip.NextId), logtype.UPGRADE_EQUIP_PET, 0, &wg)

			//Todo Update nhiệm vụ
			go user.CompleteMissionDaily(models.QDL_FORGE, &wg, int(quantity))

			wg.Wait()

			nextUEquip := models.HfUserEquip{}
			nextUEquip.Find(equip.NextId, user)

			data = iris.Map{
				"gold":       user.Gold,
				"equip":      uEquip.GetMap(),
				"next_equip": nextUEquip.GetMap(),
			}

			code, msg = 1, `Success`
		}

		c.DataResponse = iris.Map{
			"code": code,
			"msg":  msg,
			"data": data,
		}
	}
}

// /pet/combine/piece
func (c *PetController) PostCombinePiece(form formValue, ctx iris.Context) {
	if c.validToken(ctx) {
		//if true{
		user := c.User

		typePiece := uint8(util.ToInt(form("type")))
		branchPiece := form("branch")
		quantity := util.ToInt(form("quantity"))
		quantity = int(math.Max(1, float64(quantity)))

		uPiece := models.HfUserPiece{}
		uPiece, checkTypeBranch := uPiece.Get(user, branchPiece, typePiece)

		if !checkTypeBranch {
			c.DataResponse = iris.Map{"code": -1, "msg": "Level or branch invalid"}
			return
		}

		checkCombine, msg, giftPiece := uPiece.CombinePiece(quantity, user)
		if !checkCombine {
			c.DataResponse = iris.Map{"code": -1, "msg": msg}
			return
		}

		c.DataResponse = iris.Map{"code": 1, "msg": msg, "gift": giftPiece, "piece_used": uPiece.GetMap()}
	}
}

// /pet/combine/pet
func (c *PetController) PostCombinePet(form formValue, ctx iris.Context) {
	if c.validToken(ctx) {
		user := c.User
		petId := uint16(util.ToInt(form("pet_id")))

		uPet := models.HfUserPet{}
		uPet, checkUPet := uPet.Find(petId, user)

		if !checkUPet {
			c.DataResponse = iris.Map{"code": -1, "msg": "Pet invalid"}
			return
		}

		checkCombine, msg, uPet := uPet.CombinePet(user)
		if !checkCombine {
			c.DataResponse = iris.Map{"code": -1, "msg": msg}
			return
		}

		c.DataResponse = iris.Map{"code": 1, "msg": msg, "pet": uPet.GetMap()}
	}
}

// /pet/separate/piece
func (c *PetController) PostSeparatePiece(form formValue, ctx iris.Context) {
	if c.validToken(ctx) {
		//if true {
		user := c.User

		pieceQuantity := map[uint16]uint{}
		util.JsonDecodeObject(form("piece"), &pieceQuantity)

		//Kiểm tra mảnh có hợp lệ
		uPets := map[uint16]models.HfUserPet{}
		pets := map[uint16]models.HfPet{}
		for petId, quantity := range pieceQuantity {

			pet := models.HfPet{}
			pet, check := pet.Find(petId, user)

			uPet := models.HfUserPet{}
			uPet, checkPet := uPet.Find(petId, user)
			if !check || !checkPet || uPet.Piece < quantity || quantity <= 0 {
				c.DataResponse = iris.Map{"code": -1, "msg": "Piece invalid"}
				return
			}
			uPets[petId] = uPet
			pets[petId] = pet
		}

		//Lấy tỷ lệ phân tách theo cấp màu
		formula := map[uint8]int{}
		util.JsonDecodeObject(user.GetConfig("separate_piece").Value, &formula)

		//Tiến hành phân tách thành ngọc hồn
		gemSoul := 0
		var wg sync.WaitGroup
		for petId, quantity := range pieceQuantity {
			quantity := util.ToInt(quantity)

			pet := pets[petId]
			uPet := uPets[petId]

			gemSoul += quantity * formula[pet.Type]

			pieceFee := quantity * -1
			wg.Add(1)
			go uPet.SetPiece(pieceFee, 0, logtype.SEPARATE_PIECE, 0, user, &wg)

		}
		gifts := map[string]interface{}{constants.GEM_SOUL: gemSoul}

		wg.Add(1)
		go user.UpdateGifts(gifts, logtype.SEPARATE_PIECE, 0, &wg)

		wg.Wait()

		c.DataResponse = iris.Map{"code": 1, "data": iris.Map{constants.GIFT: gifts}}
	}
}

// /pet/upgrade/rune
func (c *PetController) PostUpgradeRune(form formValue, ctx iris.Context) {
	c.IsEncrypt = true
	if c.validToken(ctx) {
		//if true {
		user := c.User

		petId := uint16(util.ToInt(form("pet_id")))
		typeRune := form("type")

		pet := models.HfPet{}
		pet, check := pet.Find(petId, user)
		if pet.Type < 5 { //giới hạn pet đỏ thôi
			c.DataResponse = iris.Map{"code": -1, "msg": "Pet type invalid"}
			return
		}

		uPet := models.HfUserPet{}
		uPet, checkPet := uPet.Find(petId, user)

		petRune := models.HfPetRune{}
		petRune, checkRune := petRune.Find(typeRune, user)

		if !check || !checkPet || !checkRune || typeRune == "resist_branch" && uPet.Star < 6 || typeRune == "armor_penetration" && uPet.Star < 8 || typeRune == "critical" && uPet.Star < 11 {
			c.DataResponse = iris.Map{"code": -1, "msg": "Pet or Rune invalid"}
			return
		}

		if uPet.Stat == 0 {
			c.DataResponse = iris.Map{"code": -1, "msg": "Pet not own"}
			return
		}

		uPetRune := map[string]int{}
		util.JsonDecodeObject(uPet.Rune, &uPetRune)
		if uPetRune[typeRune] >= petRune.MaxLevel {
			c.DataResponse = iris.Map{"code": -1, "msg": "Rune's max level"}
			return
		}

		nextLevel := uPetRune[typeRune] + 1

		gemSoul := uPet.GetGemSoul(user)

		gemSoulRequest := int(math.Floor(10 * math.Pow(petRune.Material+1, float64(nextLevel-1))))
		goldRequest := int(math.Floor(200 * math.Pow(petRune.Material+1, float64(nextLevel-1))))

		if gemSoul < gemSoulRequest || user.Gold < uint(goldRequest) {
			fmt.Println("++++++++++++")
			fmt.Println(`+`, `petRune.Material`, `nextLevel`)
			fmt.Println(`+`, petRune.Material, nextLevel)
			fmt.Println(`+`, `gemSoul`, `gemSoulRequest`, `user.Gold`, `uint(goldRequest)`, `goldRequest`)
			fmt.Println(`+`, gemSoul, gemSoulRequest, user.Gold, uint(goldRequest), goldRequest)
			c.DataResponse = iris.Map{"code": -1, "msg": "Not enough Material"}
			return
		}

		//Tiền hành nâng cấp
		uPetRune[typeRune] = nextLevel
		uPet.Rune = util.JsonEndCode(uPetRune)
		var wg sync.WaitGroup
		wg.Add(3)
		go func() {
			c.DB.Save(&uPet)
			wg.Done()
		}()

		go user.SetGold(goldRequest*-1, typeRune, uint(nextLevel), logtype.UPDRADE_RUNE, 0, &wg)
		go uPet.SetGemSoul(gemSoulRequest*-1, logtype.UPDRADE_RUNE, 0, user, &wg)
		wg.Wait()

		c.DataResponse = iris.Map{"code": 1, "data": uPet.GetMap()}
	}
}

// /pet/reset/rune
func (c *PetController) PostResetRune(form formValue, ctx iris.Context) {
	c.IsEncrypt = true
	if c.validToken(ctx) {
		//if true {
		user := c.User

		petId := uint16(util.ToInt(form("pet_id")))

		uPet := models.HfUserPet{}
		uPet, checkPet := uPet.Find(petId, user)

		if !checkPet {
			c.DataResponse = iris.Map{"code": -1, "msg": "Pet invalid"}
			return
		}

		gemRequest := 200
		if user.Gem < uint(gemRequest) {
			c.DataResponse = iris.Map{"code": -1, "msg": "Not enough Gem"}
			return
		}

		uPetRune := map[string]int{}
		util.JsonDecodeObject(uPet.Rune, &uPetRune)

		goldReset := 0
		gemSoulReset := 0
		for typeRune, level := range uPetRune {
			uPetRune[typeRune] = 0
			if level <= 0 {
				continue
			}

			petRune := models.HfPetRune{}
			petRune, _ = petRune.Find(typeRune, user)
			for lv := 1; lv <= level; lv++ {

				gemSoulReset += int(math.Floor(10 * math.Pow(petRune.Material+1, float64(lv-1))))
				goldReset += int(math.Floor(200 * math.Pow(petRune.Material+1, float64(lv-1))))

			}
		}

		if gemSoulReset == 0 || goldReset == 0 {
			c.DataResponse = iris.Map{"code": -1, "msg": "Dont have Material"}
			return
		}

		gemSoulReset = int(math.Floor(float64(gemSoulReset) * 0.8))
		goldReset = int(math.Floor(float64(goldReset) * 0.8))

		//Tiền hành reset
		uPet.Rune = util.JsonEndCode(uPetRune)
		var wg sync.WaitGroup
		wg.Add(3)
		go func() {
			c.DB.Save(&uPet)
			wg.Done()
		}()

		go user.SetGem(gemRequest*-1, "", uint(uPet.PetId), logtype.RESET_RUNE, 0, &wg)

		gifts := iris.Map{constants.GOLD: goldReset, constants.GEM_SOUL: gemSoulReset}
		go user.UpdateGifts(gifts, logtype.RESET_RUNE, 0, &wg)
		wg.Wait()

		c.DataResponse = iris.Map{"code": 1, "data": iris.Map{"pet": uPet.GetMap(), constants.GIFT: gifts}}
	}
}
