package models

import (
	"GoLang/libraries/util"
	"encoding/json"
	"fmt"
	"github.com/kataras/iris"
	"sort"
)

type HfPetStar struct {
	Star    uint16 `gorm:"primary_key"`
	Type    uint8  `gorm:"primary_key"`
	Piece   uint
	GemSoul uint
	Level   uint16
	Enhance uint16
	Pet     string
	Gift    string
}

func (HfPetStar) TableName() string {
	return "hf_pet_star"
}

func (ps *HfPetStar) GetMap() iris.Map {
	return iris.Map{
		"star":     ps.Star,
		"type":     ps.Type,
		"piece":    ps.Piece,
		"gem_soul": ps.GemSoul,
		"level":    ps.Level,
		"enhance":  ps.Enhance,
		"pet":      util.JsonDecodeArray(ps.Pet),
		"gift":     util.JsonDecodeMap(ps.Gift),
	}
}

func (ps *HfPetStar) GetStar(star uint16, typePet uint8, u HfUser) (HfPetStar, bool) {
	ps.Star = star
	ps.Type = typePet
	field := fmt.Sprintf("%d_%d", typePet, star)

	count := 0
	cacheValue := u.RedisConfig.HGet(ps.TableName(), field)
	if cacheValue.Err() != nil {
		results := []HfPetStar{};
		u.DB.Where(ps).First(&results).Count(&count)

		if count != 0 {
			*ps = results[0]
			u.RedisConfig.HSet(ps.TableName(), field, util.JsonEndCode(ps))
		}
	} else {
		count = 1
		_ = json.Unmarshal([]byte(cacheValue.Val()), &ps)
	}

	return *ps, count != 0
}

func (ps *HfPetStar) CheckUpgradeStar(uPet HfUserPet, pet HfPet, uPets []interface{}, user HfUser) (bool, string) {
	//uPet, pet là pet chủ
	//uPets là danh sách pet dùng nâng cấp

	//level yêu cầu để đột phá sao
	if ps.Level > uPet.Level {
		return false, `level yêu cầu để đột phá sao`
	}

	//kiểm tra mảnh
	if ps.Piece > uPet.Piece {
		return false, `Mảnh yêu cầu để đột phá sao`
	}

	//kiểm ngọc hồn
	if int(ps.GemSoul) > uPet.GetGemSoul(user) {
		return false, `Ngọc Hồn yêu cầu để đột phá sao`
	}

	petCfg := util.JsonDecodeArray(ps.Pet)

	petsUserArr := util.ToArrayInt(uPets)
	petsUserArr = util.UniqueInt(petsUserArr)

	if len(petsUserArr) != len(petCfg) {
		return false, `len(petsUserArr) != len(petCfg)`
	}

	//Kiểm tra pet dùng nâng cấp
	for i, petId := range petsUserArr {

		petId := uint16(petId)
		//lấy pet config theo đúng thứ tự
		petRequest := util.InterfaceToMap(petCfg[i])

		//Phải khác với pet chủ
		if petId == uPet.PetId {
			return false, `Phải khác với pet chủ`
		}

		//pet cần check
		uPetCheck, checkUserPet := user.GetPet(petId)
		if !checkUserPet {
			return false, `pet cần check`
		}

		//Pet loại: Lục, Lam, Tím, Cam Đỏ
		if val, ok := petRequest["type"]; ok {
			typePet := uint8(util.ToInt(val))

			petCheck := HfPet{}
			petCheck, _ = petCheck.Find(petId, user)
			if petCheck.Type != typePet {
				return false, `Pet loại: Lục, Lam, Tím, Cam Đỏ`
			}
		}

		//Pet cùng Hệ
		if is_branch, ok := petRequest["is_branch"]; ok && util.ToInt(is_branch) == 1 {
			petCheck := HfPet{}
			petCheck, _ = petCheck.Find(petId, user)
			if petCheck.Branch != pet.Branch {
				return false, `Pet cùng Hệ`
			}
		}

		//Level1
		if val, ok := petRequest["level"]; ok {
			levelPet := uint16(util.ToInt(val))

			if uPetCheck.Level < levelPet {
				return false, `Level1`
			}
		}

		//Sao
		if val, ok := petRequest["star"]; ok {
			starPet := uint16(util.ToInt(val))

			if uPetCheck.Star < starPet {
				return false, `Sao`
			}
		}

		//Mảnh
		if val, ok := petRequest["piece"]; ok {
			piecePet := uint(util.ToInt(val))

			if uPetCheck.Piece < piecePet {
				return false, `Mảnh`
			}
		}
	}

	return true, ``
}

func (ps *HfPetStar) CheckUpgradeStar_BK(uPet HfUserPet, pet HfPet, uPets map[string]interface{}, user HfUser) bool {
	//uPet, pet là pet chủ
	//uPets là danh sách pet dùng nâng cấp

	level := HfPetLevel{}
	maxLevel := level.GetMaxLevel()

	if maxLevel[pet.Type] > uPet.Level {
		fmt.Println(1)
		return false
	}

	if ps.Piece > uPet.Piece {
		fmt.Println(2)
		return false
	}

	petCfg := util.JsonDecodeMap(ps.Pet)
	keysPet := util.MapKeys(petCfg)

	for _, key := range keysPet {
		if petsId, ok := uPets[key]; ok {

			petRequest := util.InterfaceToMap(petCfg[key])

			petsUserArr := util.ToArrayInt(petsId)
			petsUserArr = util.UniqueInt(petsUserArr)

			//kiểm tra số lương pet và có trùng trước
			if len(petsUserArr) != util.ToInt(petRequest["quantity"]) {
				fmt.Println(3)
				return false
			}

			//Kiểm tra điểu kiện của từng pet
			for i := range petsUserArr {
				petId := uint16(petsUserArr[i])

				if petId == uPet.PetId {
					fmt.Println(33)
					return false;
				}

				//pet cần check
				uPetCheck, checkUserPet := user.GetPet(petId)
				if !checkUserPet {
					fmt.Println(4)
					return false
				}

				//Pet loại: Lục, Lam, Tím, Cam Đỏ
				if val, ok := petRequest["type"]; ok {
					typePet := uint8(util.ToInt(val))

					petCheck := HfPet{}
					petCheck, _ = pet.Find(petId, user)
					if petCheck.Type != typePet {
						fmt.Println(5)
						return false;
					}
				}

				//Pet cùng Hệ
				if _, ok := petRequest["is_branch"]; ok {
					petCheck := HfPet{}
					petCheck, _ = pet.Find(petId, user)
					if petCheck.Branch != pet.Branch {
						fmt.Println(6)
						return false;
					}
				}

				//Level1
				if val, ok := petRequest["level"]; ok {
					levelPet := uint16(util.ToInt(val))

					if uPetCheck.Level < levelPet {
						fmt.Println(7)
						return false;
					}
				}

				//Sao
				if val, ok := petRequest["star"]; ok {
					starPet := uint16(util.ToInt(val))

					if uPetCheck.Star < starPet {
						fmt.Println(8)
						return false;
					}
				}
			}

		} else {
			fmt.Println(9)
			return false
		}
	}

	return true
}

func (e *HfPetStar) GetAll(u HfUser) []HfPetStar {
	results := []HfPetStar{}

	check := u.RedisConfig.Exists(e.TableName())
	if check.Val() == 0 {
		results = e.CacheAll(u)
	} else {
		cacheValue := u.RedisConfig.HGetAll(e.TableName())
		for _, val := range cacheValue.Val() {
			obj := HfPetStar{}

			_ = json.Unmarshal([]byte(val), &obj)
			results = append(results, obj)
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Type < results[j].Type || results[i].Type == results[j].Type && results[i].Star < results[j].Star
	})

	return results
}

func (e *HfPetStar) CacheAll(u HfUser) []HfPetStar {
	results := []HfPetStar{}
	u.DB.Find(&results)

	u.RedisConfig.Del(e.TableName())

	pipe := u.RedisConfig.Pipeline()
	for _, val := range results {
		field := fmt.Sprintf("%d_%d", val.Type, val.Star)
		pipe.HSet(e.TableName(), field, util.JsonEndCode(val))
	}
	_, _ = pipe.Exec()
	return results
}
