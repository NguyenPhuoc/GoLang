package models

import (
	"GoLang/libraries/constants"
	"GoLang/libraries/logtype"
	"GoLang/libraries/util"
	"database/sql"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"github.com/jinzhu/now"
	"github.com/kataras/iris"
	"sync"
	"time"
)

type HfUserEvent struct {
	UserId        string `gorm:"primary_key"`
	EventId       int    `gorm:"primary_key"`
	Type          uint
	Turn          uint
	Progress      sql.NullString
	ProgressBonus sql.NullString
	ReceiveDate   mysql.NullTime
	UpdateDate    mysql.NullTime
}

func (HfUserEvent) TableName() string {
	return "hf_user_event"
}

func (ue *HfUserEvent) Get(eventId int, u HfUser) (HfUserEvent, bool) {

	count := 0
	ue.UserId = u.UserId
	ue.EventId = eventId
	if eventId > 0 {
		u.DB.First(&ue).Count(&count)
	}

	return *ue, count != 0
}

func (ue *HfUserEvent) Init(u HfUser, group *sync.WaitGroup) {
	defer group.Done()

	group.Add(1)
	go ue.CheckinAT(u, group)

	group.Add(1)
	go ue.PVP(u, group)
}

func (HfUserEvent) CheckinAT(u HfUser, group *sync.WaitGroup) {
	defer group.Done()

	event := HfEvent{}
	event, check := event.Find(1, u)

	giftCf := map[int]interface{}{
		1: iris.Map{constants.GEM: 20000, constants.PIECE: map[int]int{3005: 50}},
		2: iris.Map{constants.GEM: 20000, constants.GOLD: 50000},
		3: iris.Map{constants.GEM: 20000, constants.PIECE: map[int]int{3020: 50}},
		4: iris.Map{constants.GEM: 25000, constants.GOLD: 75000},
		5: iris.Map{constants.GEM: 25000, constants.PIECE: map[int]int{3009: 50}},
		6: iris.Map{constants.GEM: 30000, constants.GOLD: 100000},
		7: iris.Map{constants.GEM: 50000, constants.PIECE: map[int]int{3013: 50}, constants.STONES: map[string]int{"f": 200, "e": 200, "t": 200, "w": 200, "l": 200, "d": 200}, constants.GOLD: 100000},
	}

	if check {
		ue := HfUserEvent{}
		ue, check := ue.Get(event.Id, u)
		if !check {
			progress := map[int]int{1: 0, 2: 0, 3: 0, 4: 0, 5: 0, 6: 0, 7: 0}
			ue.Progress = sql.NullString{String: util.JsonEndCode(progress), Valid: true}
		}

		if now.BeginningOfDay() != now.New(ue.ReceiveDate.Time).BeginningOfDay() {
			ue.ReceiveDate = mysql.NullTime{Time: time.Now(), Valid: true}

			progress, _ := util.JsonDecodeProgress(ue.Progress.String)

			day := (time.Now().YearDay() + 1) - event.StartDate.Time.YearDay()
			if _, ok := progress[day]; ok {
				progress[day] = 1

				ui := HfUserInbox{}
				ui.Id = util.UUID()
				ui.ReceiverId = u.UserId
				ui.SenderType = constants.INBOX_SENDER_BY_EVENT
				ui.TypeLog = logtype.GIFT_EVENT_INBOX
				ui.EventId = event.Id
				ui.KindId = day
				ui.IsReceive = 0

				title := fmt.Sprintf("Quà mở server ngày %d", day)

				ui.Title = title
				ui.Gift = sql.NullString{Valid: true, String: util.JsonEndCode(giftCf[day])}
				ui.CreatedDate = time.Now()
				u.DB.Save(&ui)

				ue.Progress = sql.NullString{String: util.JsonEndCode(progress), Valid: true}
				check = false
			}
		}

		if !check {
			u.DB.Save(&ue)
		}
	}
}

func (HfUserEvent) PVP(u HfUser, group *sync.WaitGroup) {
	defer group.Done()

	event := HfEvent{}
	event, check := event.Find(2, u)

	if check {
		ue := HfUserEvent{}
		ue, check := ue.Get(event.Id, u)

		if now.BeginningOfDay() != now.New(ue.UpdateDate.Time).BeginningOfDay() {
			ue.UpdateDate = mysql.NullTime{Time: time.Now(), Valid: true}
			ue.Turn = 0

			check = false
		}

		if !check {
			u.DB.Save(&ue)
		}
	}
}

func (HfUserEvent) GiftPVP(u HfUser, isWin bool) (bool, map[string]interface{}) {

	event := HfEvent{}
	event, check := event.Find(2, u)

	if check {
		ue := HfUserEvent{}
		ue.Get(event.Id, u)

		if now.BeginningOfDay() != now.New(ue.UpdateDate.Time).BeginningOfDay() {
			ue.UpdateDate = mysql.NullTime{Time: time.Now(), Valid: true}
			ue.Turn = 0
		}
		ue.Turn++

		u.DB.Save(&ue)
		if ue.Turn > 10 {
			return false, nil
		}

		if isWin {

			return true, map[string]interface{}{
				constants.GEM:    100,
				constants.STONES: map[string]int{"f": 10, "e": 10, "t": 10, "w": 10, "l": 10, "d": 10},
			}
		} else {

			return true, map[string]interface{}{
				constants.GEM:    50,
				constants.STONES: map[string]int{"f": 5, "e": 5, "t": 5, "w": 5, "l": 5, "d": 5},
			}
		}
	}
	return false, nil
}

func (HfUserEvent) PackageWeekMonthGrowUp(packageId string, u HfUser, group *sync.WaitGroup) {
	defer group.Done()

	switch packageId {
	case "p8": //Thẻ Tuần
		ue := HfUserEvent{}
		ue, check := ue.Get(8, u)
		//ue.ReceiveDate = Nhận tới ngày nào
		//ue.UpdateDate = Hôm nay đã nhận chưa
		if !check || ue.ReceiveDate.Valid && ue.ReceiveDate.Time.Unix() < time.Now().Unix() {
			pTime := now.EndOfDay().AddDate(0, 0, 6).Add(-1 * time.Second)
			ue.ReceiveDate = mysql.NullTime{Time: pTime, Valid: true}
		} else {
			pTime := now.New(ue.ReceiveDate.Time).AddDate(0, 0, 7)
			ue.ReceiveDate = mysql.NullTime{Time: pTime, Valid: true}
		}

		conf := HfConfig{}
		_, goldActive := conf.GetPackageGift(packageId, u)

		group.Add(3)
		go func() {
			u.DB.Save(&ue)
			group.Done()
		}()

		go u.SetGold(goldActive, packageId, 0, logtype.PAYMENT_BONUS, 8, group)
		data := iris.Map{"msg": "Active P8", "date_end": util.TimeToDateTime(ue.ReceiveDate.Time)}
		go ue.SaveLog(u, 0, 0, util.JsonEndCode(data), group)

	case "p9": //Thả Tháng
		ue := HfUserEvent{}
		ue, check := ue.Get(9, u)
		//ue.ReceiveDate = Nhận tới ngày nào
		//ue.UpdateDate = Hôm nay đã nhận chưa
		if !check || ue.ReceiveDate.Valid && ue.ReceiveDate.Time.Unix() < time.Now().Unix() {
			pTime := now.EndOfDay().AddDate(0, 0, 29).Add(-1 * time.Second)
			ue.ReceiveDate = mysql.NullTime{Time: pTime, Valid: true}
		} else {
			pTime := now.New(ue.ReceiveDate.Time).AddDate(0, 0, 30)
			ue.ReceiveDate = mysql.NullTime{Time: pTime, Valid: true}
		}

		conf := HfConfig{}
		_, goldActive := conf.GetPackageGift(packageId, u)

		group.Add(3)
		go func() {
			u.DB.Save(&ue)
			group.Done()
		}()

		go u.SetGold(goldActive, packageId, 0, logtype.PAYMENT_BONUS, 9, group)
		data := iris.Map{"msg": "Active P9", "date_end": util.TimeToDateTime(ue.ReceiveDate.Time)}
		go ue.SaveLog(u, 0, 0, util.JsonEndCode(data), group)

	case "p10":
		ue := HfUserEvent{}
		ue = ue.GetGrowUp(u)
		ue.Type = 1
		ue.ReceiveDate = mysql.NullTime{Time: time.Now(), Valid: true}
		u.DB.Save(&ue)
	}
}

func (HfUserEvent) FirstPayment(gem int, u HfUser, group *sync.WaitGroup) {
	defer group.Done()

	ue := HfUserEvent{}
	ue, check := ue.Get(7, u)
	//ue.Turn = Số Gem đã nạp tích lũy

	progress := map[uint]int{0: 2, 500: 2, 1000: 2}

	if !check {
		ue.Progress = sql.NullString{String: util.JsonEndCode(progress), Valid: true}
	}

	if ue.Turn >= 1000 { //Thoát khi đã pass
		return
	}

	util.JsonDecodeObject(ue.Progress.String, &progress)
	ue.Turn += uint(gem)
	for i, val := range progress {
		if ue.Turn >= i && val == 2 {
			progress[i] = 0
		}
	}
	ue.Progress = sql.NullString{String: util.JsonEndCode(progress), Valid: true}

	group.Add(1)
	go func() {
		u.DB.Save(&ue)
		group.Done()
	}()
}

func (ue HfUserEvent) SaveLog(u HfUser, Type, Turn uint, Data interface{}, group *sync.WaitGroup) {
	defer group.Done()
	uEventLog := HfUserEventLog{UserId: u.UserId, EventId: ue.EventId, Type: Type, KindId: Turn, Data: util.JsonEndCode(Data)}

	group.Add(1)
	go func() {
		//u.DB.Save(&uEventLog)
		group.Done()
	}()

	group.Add(1)
	go u.SaveLogMongo(uEventLog.TableName(), iris.Map{"server_id": u.ServerId, "user_id": u.UserId, "event_id": ue.EventId, "type": Type, "kind_id": Turn, "data": Data, "created_date": time.Now()}, group)
}

func (ue HfUserEvent) GetMarketBlack(u HfUser) HfUserEvent {

	//Todo reset lưu những item đã mua "trong lượt reset" đó, đang dùng đánh giấu cho 8 món chỉ được mua 1 lần duy nhất <= 8 lượt reset đó đã mua
	// limit chỉ số đã mua "trong ngày" để check random ra chợ
	// free để lưu "trong ngày" đã có random chưa và đã nhận chưa (1 món duy nhất nhưng nằm trong tệp 8 món được random)
	ue, check := ue.Get(5, u)
	if !check {
		ue.ProgressBonus = sql.NullString{String: `{"limit":{},"reset":{},"free":{}}`, Valid: true}

		shop := HfShop{}
		idItems, levelEquip, progressBonus := shop.RandomMarketBlack(u, ue.ProgressBonus.String)

		ue.Progress = sql.NullString{String: util.JsonEndCode(idItems), Valid: true}
		ue.ProgressBonus = sql.NullString{String: util.JsonEndCode(progressBonus), Valid: true}
		ue.Turn = uint(levelEquip)

		ue.ReceiveDate = mysql.NullTime{Time: time.Now(), Valid: true}
		u.DB.Save(&ue)
	}

	//qua ngày mới tính lại limit với free
	if now.New(ue.ReceiveDate.Time).BeginningOfDay() != now.BeginningOfDay() {
		ue.ReceiveDate = mysql.NullTime{Time: time.Now(), Valid: true}

		progressBonus := util.JsonDecodeMap(ue.ProgressBonus.String)
		progressBonus["limit"] = iris.Map{}
		progressBonus["free"] = iris.Map{}

		ue.ProgressBonus = sql.NullString{String: util.JsonEndCode(progressBonus), Valid: true}
		u.DB.Save(&ue)
	}

	return ue
}

func (ue HfUserEvent) GetGrowUp(u HfUser) HfUserEvent {

	ue, check := ue.Get(10, u)
	if !check {
		progress := map[int]int{5: 2, 10: 2, 15: 2, 20: 2, 25: 2, 30: 2, 35: 2, 40: 2, 45: 2, 50: 2}

		ue.Progress = sql.NullString{String: util.JsonEndCode(progress), Valid: true}
		ue.ProgressBonus = sql.NullString{String: util.JsonEndCode(progress), Valid: true}

		u.DB.Save(&ue)
	}
	level := int(u.GetLevel())
	progress := map[int]int{}
	util.JsonDecodeObject(ue.Progress.String, &progress)
	for lv, val := range progress {
		if lv <= level && val == 2 {
			progress[lv] = 0
		}
	}
	ue.Progress.String = util.JsonEndCode(progress)

	if ue.Type == 1 {
		progressBonus := map[int]int{}
		util.JsonDecodeObject(ue.ProgressBonus.String, &progressBonus)
		for lv, val := range progressBonus {
			if lv <= level && val == 2 {
				progressBonus[lv] = 0
			}
		}
		ue.ProgressBonus.String = util.JsonEndCode(progressBonus)
	}

	return ue
}

func (ue HfUserEvent) GetNewbieRollup(u HfUser) HfUserEvent {

	ue, check := ue.Get(4, u)
	if !check {
		progress := map[int]int{1: 2, 2: 2, 3: 2, 4: 2, 5: 2, 6: 2, 7: 2}

		ue.Progress = sql.NullString{String: util.JsonEndCode(progress), Valid: true}

		u.DB.Save(&ue)
	}
	day := (time.Now().Year() + time.Now().YearDay()) - (u.CreatedDate.Year() + u.CreatedDate.YearDay()) + 1
	progress := map[int]int{}
	util.JsonDecodeObject(ue.Progress.String, &progress)
	for d, val := range progress {
		if d <= day && val == 2 {
			progress[d] = 0
		}
	}
	ue.Progress.String = util.JsonEndCode(progress)

	return ue
}

func (ue HfUserEvent) GetNewbieRollupGift() (gift_cf map[int]interface{}) {
	gift_cf = map[int]interface{}{
		1: iris.Map{constants.PIECE: map[int]int{41: 20}, constants.SUMMON_BALL: map[int]int{1: 10}, constants.STONES: map[string]int{constants.FIRE: 5600}},
		2: iris.Map{constants.PIECE_GUARDIAN: map[int]int{3: 50}},
		3: iris.Map{constants.GOLD: 300000},
		4: iris.Map{constants.EQUIP: map[int]int{106: 1, 206: 1, 306: 1, 406: 1, 506: 1, 606: 1}},
		5: iris.Map{constants.SUMMON_BALL: map[int]int{1: 30, 2: 20, 3: 10}},
		6: iris.Map{constants.GEM: 1088},
		7: iris.Map{constants.PIECE: map[int]int{51: 30}},
	}
	return
}

func (ue HfUserEvent) GetRename(u HfUser) HfUserEvent {

	ue, _ = ue.Get(11, u)
	//ue.UpdateDate // update date
	//ue.Turn // số lần đổi

	return ue
}

func (ue HfUserEvent) GetRenameFee(u HfUser) int {

	if ue.Turn == 0 {
		return 0
	} else {
		return util.ToInt(u.GetConfig("rename_fee").Value)
	}
}

func (ue HfUserEvent) GetPaymentEveryday(u HfUser) HfUserEvent {

	ue, check := ue.Get(12, u)
	//ue.UpdateDate // Ngày nạp
	//ue.ReceiveDate // Ngày nhận
	//ue.Turn // Số gem đã nạp
	//ue.Progress// Mốc Nạp
	if !check || now.New(ue.UpdateDate.Time).BeginningOfDay() != now.BeginningOfDay() {
		progress := map[uint]int{200: 2, 1000: 2, 2000: 2, 5000: 2, 10000: 2, 20000: 2}

		ue.Turn = 0
		ue.Progress = sql.NullString{String: util.JsonEndCode(progress), Valid: true}
		ue.UpdateDate = mysql.NullTime{Time: time.Now(), Valid: true}
	}

	return ue
}

func (HfUserEvent) PaymentEveryday(gem int, u HfUser, group *sync.WaitGroup) {
	defer group.Done()

	ue := HfUserEvent{}
	ue = ue.GetPaymentEveryday(u)
	ue.Turn += uint(gem)

	progress := map[uint]int{}
	util.JsonDecodeObject(ue.Progress.String, &progress)
	for gem, val := range progress {
		if gem <= ue.Turn && val == 2 {
			progress[gem] = 0
		}
	}
	ue.Progress = sql.NullString{String: util.JsonEndCode(progress), Valid: true}

	u.DB.Save(&ue)
}

func (HfUserEvent) GetPaymentEverydayGift() (gift_cf map[uint]interface{}) {
	gift_cf = map[uint]interface{}{
		200: iris.Map{
			constants.GOLD: 100000,
			constants.STONES: map[string]int{
				constants.EVOLVE:  700,
				constants.FIRE:    3500,
				constants.EARTH:   3500,
				constants.THUNDER: 3500,
				constants.WATER:   3500,
				constants.LIGHT:   3500,
				constants.DARK:    3500,
			},
			constants.PIECE: map[int]int{3030: 3, 10: 3},
		},
		1000: iris.Map{
			constants.PIECE: map[int]int{3004: 3, 3030: 3, 10: 3}, constants.FLOWER_GUARDIAN: 100, constants.FRUIT_GUARDIAN: 100,
		},
		2000: iris.Map{
			constants.PIECE: map[int]int{3004: 5, 3030: 4, 10: 4}, constants.FLOWER_GUARDIAN: 130, constants.FRUIT_GUARDIAN: 130,
		},
		5000: iris.Map{
			constants.PIECE: map[int]int{3024: 15, 71: 20},
		},
		10000: iris.Map{
			constants.PIECE: map[int]int{3024: 15, 71: 20, 22: 20},
		},
		20000: iris.Map{
			constants.PIECE: map[int]int{3024: 30, 22: 30, 51: 30, 24: 30},
		},
	}
	return
}

func (ue HfUserEvent) GetPaymentAccumulate(u HfUser) HfUserEvent {

	ue, check := ue.Get(13, u)
	//ue.ReceiveDate // Ngày nhận
	//ue.Turn // Số gem đã nạp
	//ue.Progress// Mốc Nạp
	if !check {
		progress := map[int]int{500: 2, 1000: 2, 3000: 2, 6500: 2, 13500: 2, 23000: 2, 50000: 2, 70000: 2, 100000: 2, 300000: 2}

		ue.Progress = sql.NullString{String: util.JsonEndCode(progress), Valid: true}
	}

	return ue
}

func (ue HfUserEvent) GetPaymentAccumulateV2(u HfUser) (HfUserEvent, HfEvent, bool) {

	ev := HfEvent{}
	ev, checkEvent := ev.FindType("payment_accumulate_v2", u)

	if checkEvent {
		_, check := ue.Get(ev.Id, u)

		//ue.ReceiveDate // Ngày nhận
		//ue.Turn // Số gem đã nạp
		//ue.Progress// Mốc Nạp
		if !check {
			progress := map[int]int{500: 2, 1000: 2, 3000: 2, 6500: 2, 13500: 2, 23000: 2, 50000: 2, 70000: 2, 100000: 2, 300000: 2}

			ue.Progress = sql.NullString{String: util.JsonEndCode(progress), Valid: true}
		}
	}

	return ue, ev, checkEvent
}

func (HfUserEvent) PaymentAccumulate(gem int, u HfUser, group *sync.WaitGroup) {
	defer group.Done()

	ue := HfUserEvent{}
	if ue.EndTimeOpenDays(u, 7) > 0 {

		ue = ue.GetPaymentAccumulate(u)
		ue.Turn += uint(gem)

		progress := map[uint]int{}
		util.JsonDecodeObject(ue.Progress.String, &progress)
		for gem, val := range progress {
			if gem <= ue.Turn && val == 2 {
				progress[gem] = 0
			}
		}
		ue.Progress = sql.NullString{String: util.JsonEndCode(progress), Valid: true}

		u.DB.Save(&ue)
	}

	ue = HfUserEvent{}
	ue, _, check := ue.GetPaymentAccumulateV2(u)
	if check {
		ue.Turn += uint(gem)

		progress := map[uint]int{}
		util.JsonDecodeObject(ue.Progress.String, &progress)
		for gem, val := range progress {
			if gem <= ue.Turn && val == 2 {
				progress[gem] = 0
			}
		}
		ue.Progress = sql.NullString{String: util.JsonEndCode(progress), Valid: true}

		u.DB.Save(&ue)
	}
}

func (HfUserEvent) GetPaymentAccumulateGift() (gift_cf map[uint]interface{}) {
	gift_cf = map[uint]interface{}{
		500: iris.Map{
			constants.GOLD: 75000,
		},
		1000: iris.Map{
			constants.GOLD: 100000, constants.GEM: 666,
		},
		3000: iris.Map{
			constants.PIECE: map[int]int{61: 20}, constants.GOLD: 385000, constants.EQUIP: map[int]int{115: 1},
		},
		6500: iris.Map{
			constants.EQUIP: map[int]int{215: 1, 315: 1, 415: 1, 515: 1, 615: 1},
		},
		13500: iris.Map{
			constants.PIECE: map[int]int{3018: 50},
		},
		23000: iris.Map{
			constants.PIECE: map[int]int{3018: 50},
			constants.PIECE_GENERAL: []map[string]interface{}{
				{constants.BRANCH: "g", constants.TYPE: 2, constants.QUANTITY: 50},
				{constants.BRANCH: "g", constants.TYPE: 1, constants.QUANTITY: 50},
			},
		},
		50000: iris.Map{
			constants.PIECE: map[int]int{3018: 80, 61: 75},
		},
		70000: iris.Map{
			constants.PIECE: map[int]int{3006: 80}, constants.GEM: 2000,
		},
		100000: iris.Map{
			constants.PIECE: map[int]int{3006: 100},
			constants.PIECE_GENERAL: []map[string]interface{}{
				{constants.BRANCH: "g", constants.TYPE: 4, constants.QUANTITY: 60},
				{constants.BRANCH: "g", constants.TYPE: 3, constants.QUANTITY: 80},
				{constants.BRANCH: "g", constants.TYPE: 2, constants.QUANTITY: 100},
				{constants.BRANCH: "g", constants.TYPE: 1, constants.QUANTITY: 140},
			},
		},
		300000: iris.Map{
			constants.PIECE: map[int]int{3006: 295},
			constants.PIECE_GENERAL: []map[string]interface{}{
				{constants.BRANCH: "g", constants.TYPE: 4, constants.QUANTITY: 180},
				{constants.BRANCH: "g", constants.TYPE: 3, constants.QUANTITY: 240},
				{constants.BRANCH: "g", constants.TYPE: 2, constants.QUANTITY: 280},
				{constants.BRANCH: "g", constants.TYPE: 1, constants.QUANTITY: 300},
			},
		},
	}
	return
}

//ngày từ ngày đầu mở sv
func (HfUserEvent) EndTimeOpenDays(u HfUser, day int) int64 {
	sv := HfServer{}
	sv, check := sv.Find(u.ServerId, u)
	if !check {
		sv.Find(0, u)
	}
	timeOpen := sv.DateOpen.AddDate(0, 0, day-1)
	timeOpen = now.New(timeOpen).EndOfDay()

	//dateOpen := u.GetConfig("date_open_server_" + util.ToString(u.ServerId)).Value
	//timeOpen := now.MustParse(dateOpen).AddDate(0, 0, day-1)
	//timeOpen = now.New(timeOpen).EndOfDay()

	return timeOpen.Unix() - time.Now().Unix()
}

func (ue HfUserEvent) GetTurnBasic(u HfUser) HfUserEvent {

	ue, check := ue.Get(14, u)
	//ue.Turn // Số lần đã quay
	//ue.Progress// Mốc quay nhận thưởng
	if !check {
		progress := map[int]int{10: 2, 88: 2, 388: 2, 888: 2, 1388: 2, 1888: 2}

		ue.Progress = sql.NullString{String: util.JsonEndCode(progress), Valid: true}
	}

	return ue
}

func (ue HfUserEvent) GetTurnBasicV2(u HfUser) (HfUserEvent, HfEvent, bool) {

	ev := HfEvent{}
	ev, checkEvent := ev.FindType("turn_basic_v2", u)

	if checkEvent {
		_, check := ue.Get(ev.Id, u)
		//ue.Turn // Số lần đã quay
		//ue.Progress// Mốc quay nhận thưởng
		if !check {
			progress := map[int]int{10: 2, 88: 2, 388: 2, 888: 2, 1388: 2, 1888: 2}

			ue.Progress = sql.NullString{String: util.JsonEndCode(progress), Valid: true}
		}
	}

	return ue, ev, checkEvent
}

func (HfUserEvent) GetTurnBasicGift() (gift_turn []iris.Map) {
	gift_turn = []iris.Map{
		{constants.INDEX: 1, constants.PERCENT: 900, constants.GIFT: iris.Map{constants.FLOWER_GUARDIAN: 32}},
		{constants.INDEX: 2, constants.PERCENT: 900, constants.GIFT: iris.Map{constants.FRUIT_GUARDIAN: 32}},
		{constants.INDEX: 3, constants.PERCENT: 299, constants.GIFT: iris.Map{constants.SUMMON_BALL: map[int]int{3: 1}}},
		{constants.INDEX: 4, constants.PERCENT: 1100, constants.GIFT: iris.Map{constants.EQUIP: map[int]int{103: 1}}},
		{constants.INDEX: 5, constants.PERCENT: 1100, constants.GIFT: iris.Map{constants.EQUIP: map[int]int{203: 1}}},
		{constants.INDEX: 6, constants.PERCENT: 1100, constants.GIFT: iris.Map{constants.EQUIP: map[int]int{303: 1}}},
		{constants.INDEX: 7, constants.PERCENT: 1100, constants.GIFT: iris.Map{constants.EQUIP: map[int]int{403: 1}}},
		{constants.INDEX: 8, constants.PERCENT: 1100, constants.GIFT: iris.Map{constants.EQUIP: map[int]int{503: 1}}},
		{constants.INDEX: 9, constants.PERCENT: 1100, constants.GIFT: iris.Map{constants.EQUIP: map[int]int{603: 1}}},
		{constants.INDEX: 10, constants.PERCENT: 500, constants.GIFT: iris.Map{constants.PIECE_GENERAL: []map[string]interface{}{{constants.BRANCH: "g", constants.TYPE: 4, constants.QUANTITY: 1}}}},
		{constants.INDEX: 11, constants.PERCENT: 800, constants.GIFT: iris.Map{constants.PIECE_GENERAL: []map[string]interface{}{{constants.BRANCH: "g", constants.TYPE: 3, constants.QUANTITY: 1}}}},
		{constants.INDEX: 12, constants.PERCENT: 1, constants.GIFT: iris.Map{constants.PIECE: map[int]int{3015: 50}}},
	}
	return
}

func (HfUserEvent) GetTurnBasicBonus() (gift_bonus map[uint]interface{}) {
	gift_bonus = map[uint]interface{}{
		10:   iris.Map{constants.PIECE: map[int]int{12: 20}},
		88:   iris.Map{constants.PIECE: map[int]int{3009: 50}},
		388:  iris.Map{constants.PIECE_GUARDIAN: map[int]int{1: 50}},
		888:  iris.Map{constants.PIECE: map[int]int{3015: 50}},
		1388: iris.Map{constants.PIECE: map[int]int{3015: 80}},
		1888: iris.Map{constants.PIECE: map[int]int{3015: 125}},
	}
	return
}

func (ue HfUserEvent) GetGiftDiscount(u HfUser) HfUserEvent {

	ue, check := ue.Get(15, u)
	//ue.Progress// Số đã mua
	if !check {
		progress := map[int]int{1: 0, 2: 0, 3: 0, 4: 0, 5: 0, 6: 0, 7: 0}

		ue.Progress = sql.NullString{String: util.JsonEndCode(progress), Valid: true}
	}

	return ue
}
func (ue HfUserEvent) GetGiftDiscountV2(u HfUser) (HfUserEvent, HfEvent, bool) {

	ev := HfEvent{}
	ev, checkEvent := ev.FindType("gift_discount_v2", u)

	if checkEvent {
		_, check := ue.Get(ev.Id, u)
		//ue.Progress// Số đã mua
		if !check {
			progress := map[int]int{1: 0, 2: 0, 3: 0, 4: 0, 5: 0, 6: 0, 7: 0}

			ue.Progress = sql.NullString{String: util.JsonEndCode(progress), Valid: true}
		}
	}

	return ue, ev, checkEvent
}
func (HfUserEvent) GetGiftDiscountGift() (gift_cf map[int]map[string]interface{}) {
	gift_cf = map[int]map[string]interface{}{
		1: {constants.LIMIT: 1, constants.COST: 90, constants.PRICE: 1, constants.GIFT: iris.Map{constants.SUMMON_BALL: map[int]int{1: 10}}},
		2: {constants.LIMIT: 1, constants.COST: 2500, constants.PRICE: 1288, constants.GIFT: iris.Map{constants.SUMMON_BALL: map[int]int{3: 10}}},
		3: {constants.LIMIT: 1, constants.COST: 2820, constants.PRICE: 888, constants.GIFT: iris.Map{constants.PIECE: map[int]int{3023: 15}}},
		4: {constants.LIMIT: 1, constants.COST: 6250, constants.PRICE: 1888, constants.GIFT: iris.Map{constants.PIECE: map[int]int{66: 50}}},
		5: {constants.LIMIT: 1, constants.COST: 3390, constants.PRICE: 1688, constants.GIFT: iris.Map{constants.PIECE: map[int]int{3024: 15}}},
		6: {constants.LIMIT: 1, constants.COST: 6250, constants.PRICE: 1888, constants.GIFT: iris.Map{constants.PIECE: map[int]int{40: 50}}},
		7: {constants.LIMIT: 1, constants.COST: 14100, constants.PRICE: 7888, constants.GIFT: iris.Map{constants.PIECE: map[int]int{71: 75}}},
	}
	return
}

func (ue HfUserEvent) GetGiftTime(u HfUser) (HfUserEvent, HfEvent, bool) {

	ev := HfEvent{}
	ev, checkEvent := ev.FindType("gift_time", u)

	if checkEvent {
		_, check := ue.Get(ev.Id, u)

		if !check {

			ue.ReceiveDate = mysql.NullTime{Valid: true, Time: time.Now()}//thời gian bắt đầu đếm
			ue.Turn = 1 //quà đầu tiên
			u.DB.Save(&ue)
		}
	}

	return ue, ev, checkEvent
}
