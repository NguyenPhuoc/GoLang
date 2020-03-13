package models

import (
	"GoLang/config/cashshop"
	"GoLang/config/configdb"
	"GoLang/libraries/constants"
	"GoLang/libraries/logtype"
	"GoLang/libraries/msg"
	"GoLang/libraries/util"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/now"
	"github.com/kataras/iris"
	"go.mongodb.org/mongo-driver/mongo"
	"reflect"
	"strings"
	"sync"
	"time"
	"unicode"
)

type HfUser struct {
	UserId            string `gorm:"primary_key"`
	ServerId          uint
	PartnerId         int64
	UserName          string
	Code              string
	FullName          string
	Password          string
	Exp               uint64
	Gold              uint
	Gem               uint
	Power             uint16
	LastIncreasePower time.Time
	Note              sql.NullString
	LastLogin         mysql.NullTime
	NewbieStep        int
	LevelGift         uint
	AvatarId          uint16 `gorm:"default:1"`
	Access            sql.NullString
	Banned            uint8
	CreatedDate       time.Time `gorm:"default:CURRENT_TIMESTAMP"`
	DB                *gorm.DB
	RedisInfo         *redis.Client
	RedisConfig       *redis.Client
	MongoDB           *mongo.Client
}

/* Todo kiểm tra cache đi kèm user
Điểm danh 		hf_user_rollup_2019_12
Thám hiểm 		hf_user_explore
Nhiện vụ 		hf_user_mission
Nhập GiftCode	hf_gift_code.CheckInput()
*/

func (HfUser) TableName() string {
	return "hf_user"
}

func (u *HfUser) UpdateCache() {

	pipe := u.RedisInfo.Pipeline()

	s := reflect.ValueOf(u).Elem()
	typeOfT := s.Type()

	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		field := fmt.Sprintf("%s:%s", u.TableName(), typeOfT.Field(i).Name)

		pipe.HSet(u.UserId, field, util.JsonEndCode(f.Interface()))
	}

	_, _ = pipe.Exec()
}

func (u *HfUser) UpdateKey(keyCache string, val interface{}, wg *sync.WaitGroup) {
	defer wg.Done()

	field := fmt.Sprintf("%s:%s", u.TableName(), keyCache)
	u.RedisInfo.HSet(u.UserId, field, util.JsonEndCode(val))

	key := ""
	for i, val := range keyCache {
		if i > 0 && unicode.IsUpper(val) {
			key += "_" + string(val)
		} else {
			key += string(val)
		}
	}
	key = strings.ToLower(key)

	u.DB.Model(&u).Update(key, val)
}

func (u *HfUser) Get() HfUser {

	user := HfUser{}
	kLen := u.RedisInfo.HLen(u.UserId).Val()

	cacheValue := u.RedisInfo.HScan(u.UserId, 0, u.TableName()+":*", kLen)

	rVal, _ := cacheValue.Val()
	if len(rVal) > 0 {
		jVal := "{"
		for i := 0; i < len(rVal); i += 2 {
			field := strings.Replace(rVal[i], u.TableName()+":", "", 1)
			val := rVal[i+1]

			if i == 0 {
				jVal += fmt.Sprintf(`"%s":%s`, field, val)
			} else {
				jVal += fmt.Sprintf(`,"%s":%s`, field, val)
			}
		}
		jVal += "}"

		_ = json.Unmarshal([]byte(jVal), &user)
		if user.UserId == "" {
			u.DB.Where("user_id = ?", u.UserId).First(&user)
			user.RedisInfo = u.RedisInfo
			user.UpdateCache()
		}
	} else {
		u.DB.Where("user_id = ?", u.UserId).First(&user)
		user.RedisInfo = u.RedisInfo
		user.UpdateCache()
	}

	user.DB = u.DB
	user.RedisInfo = u.RedisInfo
	user.RedisConfig = u.RedisConfig
	user.MongoDB = u.MongoDB

	return user
}

func (u *HfUser) Save(wg *sync.WaitGroup) {
	defer wg.Done()
	u.UpdateCache()
	u.DB.Save(&u)
}

func (u *HfUser) SignUp() {
	u.LastIncreasePower = util.Time()
	u.AvatarId = 1

	for {
		code := util.RanCode(5)
		user := HfUser{}
		count := 0
		u.DB.Where("code = ? and server_id = ?", code, u.ServerId).First(&user).Count(&count)
		if count == 0 {
			u.Code = code
			break
		}
	}

	u.DB.Save(&u)

	var wg sync.WaitGroup
	wg.Add(1)
	go u.SetTicketArena(5, "", 0, logtype.SIGN_UP, 0, &wg)
	wg.Wait()
}

func (u *HfUser) GetPet(petId uint16) (HfUserPet, bool) {
	up := HfUserPet{PetId: petId, Stat: 1}
	up.UserId = u.UserId

	count := 0
	if petId != 0 {
		u.DB.Where(up).First(&up).Count(&count)
	}

	return up, count != 0
}

func (u *HfUser) GetStones() map[string]uint {
	us := HfUserStone{}
	usList := []HfUserStone{}
	us.UserId = u.UserId

	u.DB.Where(us).Find(&usList)

	stones := make(map[string]uint)
	stones["f"] = 0
	stones["e"] = 0
	stones["t"] = 0
	stones["w"] = 0
	stones["l"] = 0
	stones["d"] = 0
	stones["evo"] = 0

	for _, v := range usList {
		stones[v.Type] = v.Quantity
	}

	return stones
}

func (u *HfUser) SetGold(gold int, itemId interface{}, kindId uint, typeLog, eventId int, wg *sync.WaitGroup) {
	defer wg.Done()

	u.Gold = util.QuantityUint(u.Gold, gold)

	wg.Add(2)
	go u.SaveLog(typeLog, constants.GOLD, itemId, kindId, eventId, gold, uint64(u.Gold), "", wg)
	go u.UpdateKey("Gold", u.Gold, wg)
}

func (u *HfUser) SetGem(gem int, itemId interface{}, kindId uint, typeLog, eventId int, wg *sync.WaitGroup) {
	defer wg.Done()

	u.Gem = util.QuantityUint(u.Gem, gem)

	wg.Add(2)
	go u.SaveLog(typeLog, constants.GEM, itemId, kindId, eventId, gem, uint64(u.Gem), "", wg)
	go u.UpdateKey("Gem", u.Gem, wg)
}

func (u *HfUser) SetExp(exp int64, itemId uint, kindId uint, typeLog, eventId int, wg *sync.WaitGroup) {
	defer wg.Done()
	u.Exp += uint64(exp)

	wg.Add(2)
	go u.SaveLog(typeLog, constants.EXP, itemId, kindId, eventId, int(exp), uint64(u.Exp), "", wg)
	go u.UpdateKey("Exp", u.Exp, wg)
}

func (u *HfUser) SetNewbieStep(step int, wg *sync.WaitGroup) {
	defer wg.Done()
	u.NewbieStep = step

	wg.Add(1)
	go u.UpdateKey("NewbieStep", u.NewbieStep, wg)
}

func (u *HfUser) SetStones(typeStone string, quantity int, kindId uint, typeLog, eventId int, wg *sync.WaitGroup) {
	defer wg.Done()

	us := HfUserStone{Type: typeStone, UserId: u.UserId}
	if util.InArray(typeStone, us.GetTypeConfig()) {

		u.DB.Where(us).First(&us)

		us.Quantity = util.QuantityUint(us.Quantity, quantity)

		wg.Add(2)
		go u.SaveLog(typeLog, constants.STONES, typeStone, kindId, eventId, quantity, uint64(us.Quantity), "", wg)

		go func() {
			u.DB.Save(&us)
			wg.Done()
		}()
	}
}

func (u *HfUser) SetPiece(petId uint16, quantity int, kindId uint, typeLog, eventId int, wg *sync.WaitGroup) {
	defer wg.Done()

	uPet := HfUserPet{PetId: petId, UserId: u.UserId}

	count := 0
	u.DB.Where(uPet).First(&uPet).Count(&count)

	if count == 0 {
		uPet.Id = util.UUID()
	}

	wg.Add(1)
	go uPet.SetPiece(quantity, kindId, typeLog, eventId, *u, wg)
}

func (u *HfUser) SetPetEquip(equipId uint16, quantity int, kindId uint, typeLog, eventId int, wg *sync.WaitGroup) {
	defer wg.Done()

	if equipId != 0 {
		uEquip := HfUserEquip{}
		uEquip, checkUe := uEquip.Find(equipId, *u)

		if !checkUe {
			uEquip.Id = util.UUID()
		}

		uEquip.Quantity = uint16(util.QuantityUint(uEquip.Quantity, quantity))

		wg.Add(2)
		go u.SaveLog(typeLog, constants.EQUIP, uint(equipId), kindId, eventId, quantity, uint64(uEquip.Quantity), "", wg)

		go func() {
			u.DB.Save(&uEquip)
			wg.Done()
		}()
	}
}

func (u *HfUser) SetPieceGeneral(quantity int, branchPiece string, typePiece uint8, typeLog, eventId int, wg *sync.WaitGroup) {
	defer wg.Done()

	uPiece := HfUserPiece{}
	uPiece, checkTypeBranch := uPiece.Get(*u, branchPiece, typePiece)

	if checkTypeBranch {
		wg.Add(1)
		go uPiece.SetPiece(quantity, typeLog, eventId, *u, wg)
	}
}

func (u *HfUser) SetTicketArena(quantity int, itemId interface{}, kindId interface{}, typeLog, eventId int, wg *sync.WaitGroup) {
	defer wg.Done()

	uArena := HfUserArena{}
	uArena = uArena.Get(*u)

	wg.Add(1)
	go uArena.SetTicket(quantity, itemId, kindId, typeLog, eventId, *u, wg)
}

func (u *HfUser) UpdateGifts(gifts map[string]interface{}, typeGift, eventId int, wg *sync.WaitGroup) {
	defer wg.Done()

	if val, ok := gifts[constants.STONES]; ok {
		stones := util.InterfaceToMap(val)

		for sType, sQuantity := range stones {
			sTypeInt := util.ToString(sType)
			sQuantityInt := util.ToInt(sQuantity)
			if sQuantityInt > 0 {
				wg.Add(1)
				go u.SetStones(sTypeInt, sQuantityInt, 0, typeGift, eventId, wg)
			}
		}
	}

	if val, ok := gifts[constants.EXP]; ok {
		exp := int64(util.ToInt(val))
		if exp > 0 {
			wg.Add(1)
			go u.SetExp(exp, 0, 0, typeGift, eventId, wg)
		}
	}

	if val, ok := gifts[constants.GOLD]; ok {
		gold := util.ToInt(val)
		if gold > 0 {
			wg.Add(1)
			go u.SetGold(gold, "", 0, typeGift, eventId, wg)
		}
	}

	if val, ok := gifts[constants.GEM]; ok {
		gem := util.ToInt(val)
		if gem > 0 {
			wg.Add(1)
			go u.SetGem(gem, "", 0, typeGift, eventId, wg)
		}
	}

	if val, ok := gifts[constants.PIECE]; ok {
		piece := util.InterfaceToMap(val)

		for petId, quantity := range piece {
			petIdInt := uint16(util.ParseInt(petId))
			quantityInt := util.ToInt(quantity)
			if quantityInt > 0 {
				wg.Add(1)
				go u.SetPiece(petIdInt, quantityInt, 0, typeGift, eventId, wg)
			}
		}
	}

	if val, ok := gifts[constants.EQUIP]; ok {
		equip := util.InterfaceToMap(val)

		for equipId, quantity := range equip {
			equipIdInt := uint16(util.ParseInt(equipId))
			quantityInt := util.ToInt(quantity)
			if quantityInt > 0 {
				wg.Add(1)
				go u.SetPetEquip(equipIdInt, quantityInt, 0, typeGift, eventId, wg)
			}
		}
	}

	if val, ok := gifts[constants.PIECE_GENERAL]; ok {
		pieces := util.InterfaceToArray(val)

		for _, piece := range pieces {
			piece := util.InterfaceToMap(piece)

			branchPiece := util.ToString(piece["branch"])
			typePiece := uint8(util.ToInt(piece["type"]))
			quantityInt := util.ToInt(piece["quantity"])
			if quantityInt > 0 {
				wg.Add(1)
				go u.SetPieceGeneral(quantityInt, branchPiece, typePiece, typeGift, eventId, wg)
			}
		}
	}

	if val, ok := gifts[constants.TICKET_ARENA]; ok {
		ticket := util.ToInt(val)
		if ticket > 0 {
			wg.Add(1)
			go u.SetTicketArena(ticket, "", 0, typeGift, eventId, wg)
		}
	}

	if val, ok := gifts[constants.SUMMON_BALL]; ok {
		summons := util.InterfaceToMap(val)

		sum := HfUserSummon{}
		for summonId, quantity := range summons {

			summonId := uint8(util.ParseInt(summonId))
			quantity := util.ToInt(quantity)

			uSum, checkSum := sum.Get(summonId, *u)
			if checkSum && quantity > 0 {
				wg.Add(1)
				go uSum.SetQuantity(quantity, 0, typeGift, eventId, *u, wg)
			}
		}
	}

	if val, ok := gifts[constants.TICKET_MARKET_BLACK]; ok {
		quantity := util.ToInt(val)
		if quantity > 0 {
			wg.Add(1)
			go u.SetItem(constants.TICKET_MARKET_BLACK, quantity, typeGift, eventId, wg)
		}
	}

	if val, ok := gifts[constants.TICKET_GUARDIAN]; ok {
		quantity := util.ToInt(val)
		if quantity > 0 {
			wg.Add(1)
			ugi := HfUserGuardianInfo{}
			ugi.Get(*u)
			go ugi.SetTicket(quantity, "", 0, typeGift, eventId, *u, wg)
		}
	}

	if val, ok := gifts[constants.FLOWER_GUARDIAN]; ok {
		quantity := util.ToInt(val)
		if quantity > 0 {
			wg.Add(1)
			ugi := HfUserGuardianInfo{}
			ugi.Get(*u)
			go ugi.SetFlower(quantity, "", 0, typeGift, eventId, *u, wg)
		}
	}

	if val, ok := gifts[constants.FRUIT_GUARDIAN]; ok {
		quantity := util.ToInt(val)
		if quantity > 0 {
			wg.Add(1)
			ugi := HfUserGuardianInfo{}
			ugi.Get(*u)
			go ugi.SetFruit(quantity, "", 0, typeGift, eventId, *u, wg)
		}
	}

	if val, ok := gifts[constants.STONE_GUARDIAN]; ok {
		quantity := util.ToInt(val)
		if quantity > 0 {
			wg.Add(1)
			ugi := HfUserGuardianInfo{}
			ugi.Get(*u)
			go ugi.SetStone(quantity, "", 0, typeGift, eventId, *u, wg)
		}
	}

	if val, ok := gifts[constants.PIECE_GUARDIAN]; ok {
		pieces := map[uint16]int{}
		util.JsonDecodeObject(util.JsonEndCode(val), &pieces)

		for guardianId, quantity := range pieces {
			if quantity > 0 {
				wg.Add(1)
				gua := HfUserGuardian{}
				gua.Find(guardianId, *u)
				go gua.SetPiece(quantity, 0, typeGift, eventId, *u, wg)
			}
		}
	}

	if val, ok := gifts[constants.GEM_SOUL]; ok {
		quantity := util.ToInt(val)
		if quantity > 0 {
			wg.Add(1)
			go u.SetItem(constants.GEM_SOUL, quantity, typeGift, eventId, wg)
		}
	}

	if val, ok := gifts[constants.ARENA_COIN]; ok {
		quantity := util.ToInt(val)
		if quantity > 0 {
			wg.Add(1)
			uArenaShop := HfUserArenaShop{}
			uArenaShop.Get(*u)
			go uArenaShop.SetArenaCoin(quantity, "", 0, typeGift, eventId, *u, wg)
		}
	}
}

func (u *HfUser) SaveLog(typeLog int, item string, itemId interface{}, kindId interface{}, eventId int, quantity int, afterQuantity uint64, note string, wg *sync.WaitGroup) {
	defer wg.Done()

	uml := HfUserMasterLog{}
	noteLog := sql.NullString{}
	if note != "" {
		noteLog = sql.NullString{String: note, Valid: true}
	}
	itemIdS := util.ToString(itemId)
	kindIdUint := uint(util.ToInt(kindId))

	go uml.SaveLog(u, typeLog, item, itemIdS, kindIdUint, eventId, quantity, afterQuantity, noteLog)
	go func() {
		data := iris.Map{
			"server_id":      u.ServerId,
			"user_id":        u.UserId,
			"type":           typeLog,
			"item":           item,
			"item_id":        itemId,
			"kind_id":        kindId,
			"event_id":       eventId,
			"quantity":       quantity,
			"after_quantity": afterQuantity,
			"note":           note,
			"created_date":   time.Now(),
		}
		wg.Add(1)
		go u.SaveLogMongo(uml.TableName(), data, wg)
	}()
}

func (u *HfUser) SaveLogMongo(collectionName string, data iris.Map, wg *sync.WaitGroup) {
	defer wg.Done()
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("==========================")
			fmt.Println("==========================")
			fmt.Println("=========ERROR============")
			fmt.Println(err)
			fmt.Println("==========================")
			fmt.Println("==========================")
		}
	}()

	collection := u.MongoDB.Database(configdb.LOG_MONGODB).Collection(collectionName)
	if _, ok := data["data"]; ok {
		data["data"] = util.JsonDecodeMap(util.JsonEndCode(data["data"]))
	}
	_, err := collection.InsertOne(context.Background(), data)
	if err != nil {
		fmt.Println("==========================")
		fmt.Println("==========================")
		fmt.Println("=========ERROR-I==========")
		fmt.Println(err)
		fmt.Println("==========================")
		fmt.Println("==========================")

	}
}

func (u *HfUser) CompleteMissionDaily(missionId int, wg *sync.WaitGroup, quantity ...int) {
	defer wg.Done()

	uMission := HfUserMission{}
	uMission.GetDaily(*u)

	quan := 1
	if len(quantity) > 0 {
		quan = quantity[0]
	}
	uMission.CompleteDaily(missionId, *u, quan)

	uMission.CheckBigGiftDaily(*u)
}

func (u *HfUser) GetLevel() uint {
	level := HfLevel{}

	return level.GetLevel(*u)
}

func (u *HfUser) CheckLastLogin() {
	if now.New(u.LastLogin.Time).BeginningOfDay() != now.BeginningOfDay() {

		u.LastLogin = mysql.NullTime{Time: time.Now(), Valid: true}

		var wg sync.WaitGroup
		wg.Add(1)
		go u.UpdateKey("LastLogin", u.LastLogin, &wg)
		wg.Wait()
	}

	if u.Code == "" {
		for {
			code := util.RanCode(5)
			user := HfUser{}
			count := 0
			u.DB.Where("code = ? and server_id = ?", code, u.ServerId).First(&user).Count(&count)
			if count == 0 {
				u.Code = code
				var wg sync.WaitGroup
				wg.Add(1)
				go u.UpdateKey("Code", u.Code, &wg)
				wg.Wait()
				break
			}
		}
	}
}

func (u *HfUser) GetItem(item string) (HfUserItems, bool) {
	um := HfUserItems{}
	return um.Get(item, *u)
}

func (u *HfUser) SetItem(item string, quantity int, typeLog, eventId int, wg *sync.WaitGroup, timeCheck ...time.Time) {
	defer wg.Done()

	um := HfUserItems{}
	um.Get(item, *u)

	um.Quantity += quantity

	if len(timeCheck) != 0 {
		um.TimeCheck = mysql.NullTime{Time: timeCheck[0], Valid: true}
	}

	wg.Add(2)
	go u.SaveLog(typeLog, item, "", 0, eventId, quantity, uint64(um.Quantity), "", wg)
	go func() {
		u.DB.Save(&um)
		wg.Done()
	}()
}

func (u *HfUser) GetConfig(key string) HfConfig {
	con := HfConfig{}
	return con.Find(key, *u)
}

func (u *HfUser) Payment(packageId, orderId string, isWeb bool) {
	csConfig := cashshop.Config()
	cash := csConfig[packageId]

	uCashshop := HfUserCashshop{}
	uCashshop.Get(*u)

	firstPackage := map[string]uint8{}
	util.JsonDecodeObject(uCashshop.FirstPackage, &firstPackage)

	isFirstPackage := firstPackage[packageId] == 0
	if isFirstPackage {
		firstPackage[packageId] = 1
		uCashshop.FirstPackage = util.JsonEndCode(firstPackage)
	}

	isFirstPayment := uCashshop.FirstPayment == 0
	if isFirstPayment {
		uCashshop.FirstPayment = 1
	}

	dataNotify := iris.Map{"package_id": cash.PackageId, constants.GEM: cash.Gem}

	var wg sync.WaitGroup
	wg.Add(1)
	go u.SetGem(cash.Gem, packageId, 0, logtype.PAYMENT, 0, &wg)
	wg.Wait()

	if isFirstPackage && cash.GemFirst > 0 {
		wg.Add(1)
		go u.SetGem(cash.GemFirst, packageId, 0, logtype.PAYMENT_BONUS, 0, &wg)
		dataNotify["gem_bonus"] = cash.GemFirst
	}

	wg.Add(1)
	go func() {
		u.DB.Save(&uCashshop)
		wg.Done()
	}()

	order := sql.NullString{String: orderId, Valid: true}
	if orderId == "" {
		order = sql.NullString{}
	}

	uPayLog := HfUserPaymentLog{Id: util.UUID(), UserId: u.UserId, PackageId: packageId, OrderId: order}
	if isFirstPackage {
		uPayLog.IsFirst = 1
	}

	wg.Add(1)
	go func() {
		u.DB.Save(&uPayLog)
		wg.Done()
	}()

	wg.Add(4)
	ue := HfUserEvent{}
	go ue.PackageWeekMonthGrowUp(packageId, *u, &wg)
	go ue.FirstPayment(cash.Gem, *u, &wg)
	go ue.PaymentEveryday(cash.Gem, *u, &wg)
	go ue.PaymentAccumulate(cash.Gem, *u, &wg)

	wg.Wait()

	if isWeb {

		conf := HfConfig{}
		if packageId == "p8" || packageId == "p9" {
			_, goldActive := conf.GetPackageGift(packageId, *u)
			dataNotify[constants.GOLD] = goldActive
		}

		infoNotify := iris.Map{
			"server_id": u.ServerId,
			"user_id":   u.UserId,
			"data":      dataNotify,
		}
		u.RedisInfo.Publish(msg.CHANNEL_PAYMENT, util.JsonEndCode(infoNotify))
	}
}
