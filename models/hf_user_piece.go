package models

import (
	"GoLang/libraries/constants"
	"GoLang/libraries/logtype"
	"GoLang/libraries/util"
	"github.com/kataras/iris"
	"math/rand"
	"sync"
)

/*Mảnh chung, mảnh riêng
Chung: branch = g
Riêng: branch = d, l, w, t, e, f*/
type HfUserPiece struct {
	Id       string `gorm:"primary_key"`
	UserId   string
	Branch   string
	Type     uint8
	Quantity uint
}

func (HfUserPiece) TableName() string {
	return "hf_user_piece"
}

func (up *HfUserPiece) GetMap() iris.Map {
	return iris.Map{
		"branch":   up.Branch,
		"type":     up.Type,
		"quantity": up.Quantity,
	}
}

func (HfUserPiece) GetTypeConfig() []uint8 {
	return []uint8{1, 2, 3, 4, 5}
}

func (HfUserPiece) GetBranchConfig() []string {
	return []string{"g", "d", "l", "w", "t", "e", "f"}
}

func (up *HfUserPiece) Get(u HfUser, branchPiece string, typePiece uint8) (HfUserPiece, bool) {

	checkBranch := util.InArray(branchPiece, up.GetBranchConfig())
	checkType := util.InArray(typePiece, up.GetTypeConfig())

	if checkBranch && checkType {

		up.UserId = u.UserId
		up.Branch = branchPiece
		up.Type = typePiece

		count := 0
		u.DB.Where(up).Find(&up).Count(&count)

		if count == 0 {
			up.Id = util.UUID()
			u.DB.Save(&up)
		}

		return *up, true
	}

	return *up, false
}

func (up *HfUserPiece) SetPiece(piece int, typeLog, eventId int, u HfUser, wg *sync.WaitGroup) {
	defer wg.Done()

	up.Quantity = util.QuantityUint(up.Quantity, piece)

	wg.Add(2)
	go u.SaveLog(typeLog, constants.PIECE_GENERAL, up.Branch, up.Type, eventId, piece, uint64(up.Quantity), "", wg)

	go func() {
		u.DB.Save(&up)
		wg.Done()
	}()
}

func (HfUserPiece) GetPieces(u HfUser) string {
	pieces := []struct {
		Branch   string `json:"branch"`
		Type     uint8  `json:"type"`
		Quantity uint   `json:"quantity"`
	}{}
	u.DB.Raw(`SELECT branch, type, quantity FROM hf_user_piece WHERE user_id = ?;`,u.UserId).Scan(&pieces)

	return util.JsonEndCode(pieces)
}

func (up *HfUserPiece) CombinePiece(quantity int, u HfUser) (bool, string, interface{}) {

	//Kiểm tra số lượng theo cấp màu (điều kiện check phải đủ để thành 1 bộ cho pet)
	petStar := HfPetStar{}
	petStar.GetStar(0, up.Type, u)

	//số mảnh yêu cầu
	quanRequest := uint(int(petStar.Piece) * quantity)
	if quanRequest > up.Quantity {
		return false, "Quantity not enough", nil
	}

	//Lấy danh sách pet có thể xuất hiện
	pet := HfPet{}
	allPets := pet.GetAll(u)
	petRans := []HfPet{}

	for _, pet := range allPets {
		//Nếu là mảnh chung => chỉ kiểm tra type
		if up.Branch == "g" {
			if pet.Type == up.Type {
				petRans = append(petRans, pet)
			}
		} else {
			if pet.Type == up.Type && pet.Branch == up.Branch {
				petRans = append(petRans, pet)
			}
		}
	}

	if len(petRans) == 0 {
		return false, "Pet does not exist", nil
	}

	//Lấy tổng độ hiếm
	totalRarityPet := 0
	for _, pet := range petRans {
		totalRarityPet += int(pet.Rarity)
	}

	//random theo số bộ mảnh có
	gifts := map[string]interface{}{}
	for i := 0; i < quantity; i++ {
		//Random theo độ hiếm và nhận nguyên bộ ghep thành pet
		ranRarity := rand.Intn(totalRarityPet)
		perRarity := 0

		for _, pet := range petRans {
			perRarity += int(pet.Rarity)
			if perRarity > ranRarity {

				giftPiece := map[string]interface{}{
					constants.PIECE: map[uint16]uint{
						pet.Id: petStar.Piece,
					},
				}
				gifts = util.MergeGift(gifts, giftPiece)

				break
			}
		}
	}

	//trừ mảnh đã dùng và update mảnh(gift)
	pieceFee := int(quanRequest) * -1

	var wg sync.WaitGroup
	wg.Add(2);
	go up.SetPiece(pieceFee, logtype.PIECE_COMBINE_PIECE, 0, u, &wg)
	go u.UpdateGifts(gifts, logtype.PIECE_COMBINE_PIECE, 0, &wg)
	wg.Wait()

	return true, "Success", gifts
}
