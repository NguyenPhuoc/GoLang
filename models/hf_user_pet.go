package models

import (
	"GoLang/libraries/constants"
	"GoLang/libraries/logtype"
	"GoLang/libraries/util"
	"github.com/jinzhu/gorm"
	"github.com/kataras/iris"
	"sync"
)

type HfUserPet struct {
	Id      string `gorm:"primary_key"`
	UserId  string
	PetId   uint16
	Stat    uint8
	Piece   uint
	Level   uint16
	Evolve  uint16
	Star    uint16
	Enhance uint16
	Skill   string `gorm:"default:'{\"1\": 1, \"2\": 1, \"3\": 1, \"4\": 1}'"`
	Rune    string `gorm:"default:'{\"hp\":0,\"damage\":0,\"resist_branch\":0,\"armor\":0,\"critical\":0,\"armor_penetration\":0}'"`
}
type petEquip struct {
	HfUserPet
	Equip map[uint8]uint16
}

func (HfUserPet) TableName() string {
	return "hf_user_pet"
}

//Số mảnh yêu cầu để ghép thành pet tương ứng cấp màu
func (HfUserPet) GetTypeCombineConfig() map[uint8]int {
	return map[uint8]int{1: 5, 2: 10, 3: 20, 4: 30, 5: 50}
}

func (up *HfUserPet) Find(petId uint16, u HfUser) (HfUserPet, bool) {
	up.UserId = u.UserId
	up.PetId = petId

	count := 0
	if petId != 0 {
		u.DB.Where(up).First(&up).Count(&count)
	}

	return *up, count != 0
}

func (up *HfUserPet) GetMap() iris.Map {
	return iris.Map{
		"id":      up.PetId,
		"stat":    up.Stat,
		"piece":   up.Piece,
		"level":   up.Level,
		"evolve":  up.Evolve,
		"star":    up.Star,
		"enhance": up.Enhance,
		"skill":   util.JsonDecodeMap(up.Skill),
		"rune":    util.JsonDecodeMap(up.Rune),
	}
}

func (pe *petEquip) GetMap() iris.Map {
	uPetEquip := pe.HfUserPet.GetMap()
	uPetEquip[constants.EQUIP] = map[uint8]uint16{
		1: pe.Equip[1],
		2: pe.Equip[2],
		3: pe.Equip[3],
		4: pe.Equip[4],
		5: pe.Equip[5],
		6: pe.Equip[6],
	}

	return uPetEquip
}

func (up *HfUserPet) PetEquip() petEquip {
	pe := petEquip{HfUserPet: *up}

	pe.Equip = map[uint8]uint16{1: 0, 2: 0, 3: 0, 4: 0, 5: 0, 6: 0}

	return pe
}

func (up *HfUserPet) SetPiece(piece int, kindId uint, typeLog, eventId int, u HfUser, wg *sync.WaitGroup) {
	defer wg.Done()

	up.Piece = util.QuantityUint(up.Piece, piece)

	wg.Add(2)
	go u.SaveLog(typeLog, constants.PIECE, uint(up.PetId), kindId, eventId, piece, uint64(up.Piece), "", wg)

	go func() {
		u.DB.Save(&up)
		wg.Done()
	}()
}

func (up *HfUserPet) GetPetEquip(u HfUser) []iris.Map {

	results := []struct {
		HfUserPet
		EquipType uint8
		EquipId   uint16
	}{}

	u.DB.Raw(`SELECT p.*, pe.equip_type, e.equip_id FROM hf_user_pet p
	LEFT JOIN hf_user_pet_equip pe ON p.id = pe.user_pet_id
	LEFT JOIN hf_user_equip e ON pe.user_equip_id = e.id
	WHERE p.user_id = ?`, u.UserId).Scan(&results)

	petEquip := map[uint16]petEquip{}
	for _, val := range results {
		uPet := val.HfUserPet.PetEquip()

		if pe, ok := petEquip[val.PetId]; ok {
			uPet = pe
		}

		uPet.Equip[val.EquipType] = val.EquipId

		petEquip[val.PetId] = uPet
	}

	petEquipReturn := make([]iris.Map, len(petEquip))
	i := 0
	for _, val := range petEquip {
		petEquipReturn[i] = val.GetMap()
		i++
	}

	return petEquipReturn
}

func (up *HfUserPet) GetPetEquipMap(db *gorm.DB) iris.Map {

	results := []struct {
		HfUserPet
		EquipType uint8
		EquipId   uint16
	}{}

	db.Raw(`SELECT p.*, pe.equip_type, e.equip_id FROM hf_user_pet p
	LEFT JOIN hf_user_pet_equip pe ON p.id = pe.user_pet_id
	LEFT JOIN hf_user_equip e ON pe.user_equip_id = e.id
	WHERE p.user_id = ? AND p.pet_id = ?`, up.UserId, up.PetId).Scan(&results)

	petEquip := map[uint16]petEquip{}
	for _, val := range results {
		uPet := val.HfUserPet.PetEquip()

		if pe, ok := petEquip[val.PetId]; ok {
			uPet = pe
		}

		uPet.Equip[val.EquipType] = val.EquipId

		petEquip[val.PetId] = uPet
	}

	pEquip := petEquip[up.PetId];
	return pEquip.GetMap()
}

func (up *HfUserPet) GetMaterialUpgraded(db *gorm.DB) iris.Map {

	result := struct {
		PetId     uint16
		Type      string
		Piece     float64
		StoneType float64
		Gold      float64
		StoneEvo  float64
	}{}
	ratio := 0.8

	db.Raw(`SELECT t.type, t.pet_id,
	(SELECT sum(ps.piece) FROM hf_pet_star ps WHERE ps.star <= t.star AND ps.star!=0) piece,
	(SELECT SUM(pl.stone) FROM hf_pet_level pl WHERE pl.level <= t.LEVEL) stone_type,
	(SELECT SUM(pl.gold) FROM hf_pet_level pl WHERE pl.level <= t.LEVEL) +
	(SELECT SUM(pe.gold) FROM hf_pet_evolve pe WHERE pe.evolve <= t.evolve) gold,
	(SELECT SUM(pe.stone) FROM hf_pet_evolve pe WHERE pe.evolve <= t.evolve) stone_evo
	FROM (
		SELECT up.level, up.evolve, up.star, p.type, up.pet_id FROM hf_user_pet up JOIN hf_pet p ON up.pet_id = p.id
		WHERE up.user_id = ? AND up.pet_id = ?
	) t`, up.UserId, up.PetId).Scan(&result)

	material := iris.Map{}
	if result.Gold > 0 {
		material[constants.GOLD] = int(result.Gold * ratio)
	}

	if result.StoneEvo > 0 {
		if val, ok := material[constants.STONES]; ok {
			stones := util.InterfaceToMap(val)

			stones[constants.EVOLVE] = int(result.StoneEvo * ratio)
			material[constants.STONES] = stones
		} else {
			material[constants.STONES] = iris.Map{
				constants.EVOLVE: int(result.StoneEvo * ratio),
			}
		}
	}

	if result.StoneType > 0 {
		if val, ok := material[constants.STONES]; ok {
			stones := util.InterfaceToMap(val)

			stones[result.Type] = int(result.StoneType * ratio)
			material[constants.STONES] = stones
		} else {
			material[constants.STONES] = iris.Map{
				result.Type: int(result.StoneType * ratio),
			}
		}
	}

	if result.Piece > 0 {
		petId := util.ToString(result.PetId)

		material[constants.PIECE] = iris.Map{
			petId: int(result.Piece * ratio),
		}
	}

	return material
}

func (up *HfUserPet) CombinePet(u HfUser) (bool, string, HfUserPet) {

	if up.Stat == 1 {
		return false, "Pet has combine", *up
	}

	pet := HfPet{}
	pet, checkPet := pet.Find(up.PetId, u)
	if !checkPet {
		return false, "Pet does not exist", *up
	}

	//Kiểm tra số lượng theo cấp màu
	typeCombineCf := up.GetTypeCombineConfig()
	quanRequest := uint(typeCombineCf[pet.Type])
	if quanRequest > up.Piece {
		return false, "Quantity piece not enough", *up
	}

	//Tiến hành hợp pet
	up.Stat = 1

	//trừ mảnh đã dùng
	pieceFee := int(quanRequest) * -1

	var wg sync.WaitGroup
	wg.Add(1);
	go up.SetPiece(pieceFee, 0, logtype.PIECE_COMBINE_PET, 0, u, &wg)
	wg.Wait()

	return true, "Success", *up
}

func (up *HfUserPet) GetGemSoul(u HfUser) int {
	uItem, _ := u.GetItem(constants.GEM_SOUL)
	return uItem.Quantity
}

func (up *HfUserPet) SetGemSoul(quantity, typeLog, eventId int, u HfUser, wg *sync.WaitGroup) {

	go u.SetItem(constants.GEM_SOUL, quantity, typeLog, eventId, wg)
}
