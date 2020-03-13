package models

import (
	"GoLang/libraries/constants"
	"GoLang/libraries/logtype"
	"GoLang/libraries/util"
	"github.com/go-sql-driver/mysql"
	"github.com/kataras/iris"
	"math"
	"math/rand"
	"sync"
	"time"
)

type HfUserSummon struct {
	UserId   string `gorm:"primary_key"`
	SummonId uint8  `gorm:"primary_key"`//Todo 1:lục 2:tím 3:đỏ
	Quantity uint
	TurnFree mysql.NullTime
}

func (HfUserSummon) TableName() string {
	return "hf_user_summon"
}

func (us *HfUserSummon) GetMap(u HfUser) iris.Map {
	timeCount := float64(-1)

	sum := HfSummon{}
	sum.Find(us.SummonId, u)

	if sum.TimeFree > 0 {
		timeCount = math.Max(0, float64(sum.TimeFree*60)-float64(time.Now().Unix()-us.TurnFree.Time.Unix()))
	}

	return iris.Map{
		"quantity":    us.Quantity,
		"turn_free":   timeCount,
		"turn_fee_1":  sum.TurnFee1,
		"turn_fee_10": sum.TurnFee10,
	}
}

func (us *HfUserSummon) Get(summonId uint8, u HfUser) (HfUserSummon, bool) {
	us.UserId = u.UserId
	us.SummonId = summonId

	sum := HfSummon{}
	sum, checkSum := sum.Find(summonId, u)

	if checkSum {
		count := 0
		u.DB.First(&us).Count(&count)

		if count == 0 {
			if sum.TimeFree > 0 {
				us.TurnFree = mysql.NullTime{Time: time.Now(), Valid: true}
			}
			u.DB.Save(&us)
		} else if us.TurnFree.Valid == false && sum.TimeFree > 0 {

			us.TurnFree = mysql.NullTime{Time: time.Now(), Valid: true}
			u.DB.Save(&us)
		}
	}

	return *us, checkSum
}

func (us *HfUserSummon) GetLogType(turnType string) int {
	typesLog := map[string]map[uint8]int{
		"turn": {
			1: logtype.GIFT_SUMMON1_TURN,
			2: logtype.GIFT_SUMMON2_TURN,
			3: logtype.GIFT_SUMMON3_TURN,
		},
		"free": {
			1: logtype.GIFT_SUMMON1_FREE,
			2: logtype.GIFT_SUMMON2_FREE,
			3: logtype.GIFT_SUMMON3_FREE,
		},
		"gem": {
			1: logtype.BUY_GIFT_SUMMON1,
			2: logtype.BUY_GIFT_SUMMON2,
			3: logtype.BUY_GIFT_SUMMON3,
		},
	}

	return typesLog[turnType][us.SummonId]
}

func (us *HfUserSummon) SetQuantity(quantity int, kindId uint, typeLog, eventId int, u HfUser, wg *sync.WaitGroup) {
	defer wg.Done()

	us.Quantity = util.QuantityUint(us.Quantity, quantity)

	wg.Add(2)
	go u.SaveLog(typeLog, constants.SUMMON_BALL, us.SummonId, kindId, eventId, quantity, uint64(us.Quantity), "", wg)

	go func() {
		u.DB.Save(&us)
		wg.Done()
	}()
}

func (us *HfUserSummon) Random(quantity int, u HfUser) (map[string]interface{}, []interface{}) {
	sum := HfSummon{}
	sum.Find(us.SummonId, u)

	totalRatioBranch := 0
	branchRandom := util.JsonDecodeArray(sum.BranchRandom)
	for _, item := range branchRandom {
		item := util.InterfaceToMap(item)
		if per, ok := item[constants.PERCENT]; ok {
			totalRatioBranch += util.ToInt(per)
		}
	}

	totalRatioType := 0
	typeRandom := util.JsonDecodeArray(sum.TypeRandom)
	for _, item := range typeRandom {
		item := util.InterfaceToMap(item)
		if per, ok := item[constants.PERCENT]; ok {
			totalRatioType += util.ToInt(per)
		}
	}

	pet := HfPet{}
	allPet := pet.GetAll(u)

	gifts := map[string]interface{}{}
	giftsReturn := []interface{}{}
	for i := 0; i < quantity; i++ {

		petRans := []HfPet{}
		for {//Todo Lặp lại nếu không random ra pet
			ranBranch := rand.Intn(totalRatioBranch)
			ranType := rand.Intn(totalRatioType)

			var typePet uint8
			var branchPet string

			percent := 0
			for _, item := range branchRandom {
				item := util.InterfaceToMap(item)
				if per, ok := item[constants.PERCENT]; ok {
					percent += util.ToInt(per)
					if percent > ranBranch {
						branchPet = util.ToString(item["branch"])
						break
					}
				}
			}

			percent = 0
			for _, item := range typeRandom {
				item := util.InterfaceToMap(item)
				if per, ok := item[constants.PERCENT]; ok {
					percent += util.ToInt(per)
					if percent > ranType {
						typePet = uint8(util.ToInt(item["type"]))
						break
					}
				}
			}
			for _, pet := range allPet {
				if pet.TypeRandom == typePet && pet.Branch == branchPet {
					petRans = append(petRans, pet)
				}
			}

			if len(petRans) > 0 {
				break
			}
		}

		totalRarityPet := 0
		for _, pet := range petRans {
			totalRarityPet += int(pet.Rarity)
		}
		ranRarity := rand.Intn(totalRarityPet)
		perRarity := 0
		for _, pet := range petRans {
			perRarity += int(pet.Rarity)
			if perRarity > ranRarity {

				//petStar := HfPetStar{}
				//petStar.GetStar(0, pet.Type, u)

				giftPiece := map[string]interface{}{
					constants.PIECE: map[uint16]uint{
						//pet.Id: petStar.Piece,
						pet.Id: 5,
					},
				}
				gifts = util.MergeGift(gifts, giftPiece)
				giftsReturn = append(giftsReturn, giftPiece)

				break
			}
		}

	}
	return gifts, giftsReturn
}
