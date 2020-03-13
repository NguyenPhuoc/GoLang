package models

import (
	"GoLang/libraries/constants"
	"GoLang/libraries/logtype"
	"GoLang/libraries/util"
	"database/sql"
	"github.com/jinzhu/gorm"
	"github.com/kataras/iris"
	"math"
	"math/rand"
	"sync"
	"time"
)

func (HfUserArena) GetPriceTicket(quantity int) int {
	//Giá của mỗi ticket là 50 gem
	return quantity * 50
}

type HfUserArena struct {
	UserId           string `gorm:"primary_key"`
	FullName         string `gorm:"default:null"`
	IsBoss           uint8
	RivalId          string `gorm:"default:null"`
	Ticket           uint
	Battle           uint
	MatchHistory     string
	LineUp           sql.NullString
	Elo              int `gorm:"default:1000"`
	PowerPoint       int
	BossId           int
	Armor            float64
	ArmorPenetration float64
	Critical         float64
	ResistBranch     string `gorm:"default:'{\"m\":0,\"w\":0,\"t\":0,\"f\":0,\"l\":0,\"d\":0}'"`
	Hp               int
	Damage           int
	Star             int
	Level            int
	Iq               int
	Skill            sql.NullString
	Support          sql.NullString
	Guardian         sql.NullString
	LastLineUp       sql.NullString
	HitDaily         int
	HitWeekly        int
}

func (HfUserArena) TableName() string {
	return "hf_user_arena"
}

func (ua *HfUserArena) GetMap() iris.Map {
	return iris.Map{
		"user_id":              ua.UserId,
		"full_name":            ua.FullName,
		"rival_id":             ua.RivalId,
		constants.TICKET_ARENA: ua.Ticket,
		"battle":               ua.Battle,
		"line_up":              util.JsonDecodeMap(ua.LineUp.String),
		"elo":                  ua.Elo,
		"power_point":          ua.PowerPoint,
		"boss_id":              ua.BossId,
		"armor":                ua.Armor,
		"armor_penetration":    ua.ArmorPenetration,
		"critical":             ua.Critical,
		"damage":               ua.Damage,
		"star":                 ua.Star,
		"hp":                   ua.Hp,
		"iq":                   ua.Iq,
		"is_boss":              ua.IsBoss,
		"level":                ua.Level,
		"resist_branch":        util.JsonDecodeMap(ua.ResistBranch),
		"support":              util.JsonDecodeMap(ua.Support.String),
		"guardian":             util.JsonDecodeMap(ua.Guardian.String),
		"skill":                util.JsonDecodeMap(ua.Skill.String),
	}
}

func (ua *HfUserArena) Get(u HfUser) HfUserArena {
	ua.UserId = u.UserId

	count := 0
	u.DB.First(&ua).Count(&count)

	if count == 0 {
		ua.FullName = u.FullName
		ua.MatchHistory = "{}"
		ua.Elo = 1000
		u.DB.Create(&ua)
	}

	return *ua
}

func (ua *HfUserArena) Find(userId string, db *gorm.DB) (HfUserArena, bool) {
	ua.UserId = userId

	count := 0
	if userId != "" {
		db.First(&ua).Count(&count)
	}

	return *ua, count != 0
}

func (ua *HfUserArena) SetTicket(quantity int, itemId interface{}, kindId interface{}, typeLog, eventId int, u HfUser, wg *sync.WaitGroup) {
	defer wg.Done()

	ua.Ticket = util.QuantityUint(ua.Ticket, quantity)

	wg.Add(2)
	go u.SaveLog(typeLog, constants.TICKET_ARENA, itemId, kindId, eventId, quantity, uint64(ua.Ticket), "", wg)

	go func() {
		u.DB.Save(&ua)
		wg.Done()
	}()
}

func (ua *HfUserArena) GetRankAndTop(db *gorm.DB) (int, string) {

	tops := []struct {
		UserId     string `json:"user_id"`
		FullName   string `json:"full_name"`
		Level      int    `json:"level"`
		Elo        int    `json:"elo"`
		PowerPoint int    `json:"power_point"`
		Top        int    `json:"top"`
		AvatarId   int    `json:"avatar_id"`
	}{}
	//db.Raw(`SELECT user_id, full_name, level, elo, power_point FROM hf_user_arena WHERE is_boss = 0 ORDER BY elo DESC LIMIT 100;`).Scan(&tops)
	//db.Raw(`SELECT are.user_id, full_name, level, elo, power_point, ava.avatar_id FROM hf_user_arena are
	//			LEFT JOIN (SELECT user_id, MAX(avatar_id) avatar_id FROM hf_user_avatar WHERE used = 1 GROUP BY user_id) ava
	//			ON ava.user_id = are.user_id
	//			WHERE is_boss = 0 AND power_point != 0
	//			ORDER BY elo DESC LIMIT 100;`).Scan(&tops)
	db.Raw(`SELECT are.user_id, are.full_name, level, elo, power_point, ava.avatar_id FROM hf_user_arena are 
				LEFT JOIN hf_user ava 
				ON ava.user_id = are.user_id 
				WHERE is_boss = 0 AND power_point != 0 
				ORDER BY elo DESC LIMIT 100;`).Scan(&tops)

	myRank := 0;
	for i, val := range tops {
		val.Top = i + 1
		tops[i].Top = val.Top

		if val.UserId == ua.UserId {
			myRank = val.Top
		}
	}
	if myRank == 0 {
		db.Model(&HfUserArena{}).Where("elo > ?", ua.Elo).Count(&myRank)
		myRank++;
	}

	return myRank, util.JsonEndCode(tops)
}

func (ua *HfUserArena) RandomRival(u HfUser) string {

	results := []struct {
		UserId string
	}{}

	//15 user lớn hơn và bé hơn elo bản thân
	u.DB.Raw(`SELECT * FROM (SELECT user_id, elo FROM hf_user_arena WHERE (elo >= ? AND user_id != ? and line_up is not null and line_up != '' and line_up != '{}') OR (is_boss = 1 AND elo >= ?) ORDER BY elo ASC LIMIT 15) t1
		UNION 
		SELECT * FROM (SELECT user_id, elo FROM hf_user_arena WHERE (elo < ? AND user_id != ? and line_up is not null and line_up != '' and line_up != '{}') OR (is_boss = 1 AND elo < ?) ORDER BY elo DESC LIMIT 15) t2`,
		ua.Elo, ua.UserId, ua.Elo,
		ua.Elo, ua.UserId, ua.Elo).Scan(&results)

	ua.Battle++ //tăng lượt đánh

	matchHistory := util.JsonDecodeMap(ua.MatchHistory)
	//kiểm tra loại bỏ đối thủ đã gặp xa hơn 5 trận trước từ trong lịch sử
	for id, bat := range matchHistory {
		bat := uint(bat.(float64)) + 5 //Giới hạn số trận để gặp lại đối thủ
		if bat <= ua.Battle {
			delete(matchHistory, id)
		}
	}

	//Lấy danh sách hợp lệ để random
	ranUserId := make(map[string]interface{})
	for _, val := range results {
		if _, ok := matchHistory[val.UserId]; !ok {
			ranUserId[val.UserId] = 0;
		}
	}

	userIds := util.MapKeys(ranUserId)
	rivalId := ""
	if len(userIds) != 0 {
		n := rand.Int() % len(userIds)
		rivalId = userIds[n]

		//lưu lịch sử trận đánh và đối thủ
		ua.RivalId = rivalId
		//fmt.Println(matchHistory)
		matchHistory[rivalId] = ua.Battle

		ua.MatchHistory = util.JsonEndCode(matchHistory)

		u.DB.Save(&ua)
	}

	return rivalId
}

func (ua *HfUserArena) EndgameWinLose(uRival *HfUserArena, isWin bool, data string, u HfUser) (int, interface{}) {

	elo := float64(0);
	eloR := float64(0);
	if isWin {
		//Todo Elo cộng =MIN(45;MAX(1;30 - (elo bản thân - elo đối thủ)/100) )
		elo = math.Min(45, math.Max(1, float64(30-(ua.Elo-uRival.Elo)/100)))
		//Todo Đối thủ bị trừ 50% Elo cộng(làm tròn lên)
		eloR = math.Round(elo / 2 * -1)
	} else {
		//Todo Elo thua = MIN(-1;MAX(-45;(elo đối thủ - elo bản thân)/100 -30)))
		elo = math.Min(-1, math.Max(-45, float64((uRival.Elo-ua.Elo)/100-30)))
		//Todo Đối thủ được cộng 50% Elo trừ(làm tròn lên)
		eloR = math.Round(elo / 2 * -1)
	}
	ua.Elo += int(elo)
	uRival.Elo += int(eloR)

	ua.RivalId = ""

	//save log
	var wg sync.WaitGroup
	ual := HfUserArenaLog{UserAtt: ua.UserId, UserDef: uRival.UserId, EloAtt: int(elo), EloDef: int(eloR), EloAttAfter: ua.Elo, EloDefAfter: uRival.Elo, Data: data}
	//go u.DB.Create(&ual)
	wg.Add(1)
	go u.SaveLogMongo(ual.TableName(), iris.Map{"server_id": u.ServerId, "user_att": ua.UserId, "user_def": uRival.UserId, "elo_att": int(elo), "elo_def": int(eloR), "elo_att_after": ua.Elo, "elo_def_after": uRival.Elo, "data": util.JsonDecodeMap(data), "created_date": time.Now()}, &wg)

	u.DB.Save(&ua)
	if uRival.IsBoss == 0 {
		u.DB.Save(&uRival)
	}

	//update gift
	wg.Add(2)
	gift := ua.GiftEndgame(isWin)
	go u.UpdateGifts(gift, logtype.GIFT_ENDGAME_ARENA, 0, &wg)
	//Todo Update nhiệm vụ
	go u.CompleteMissionDaily(QDL_ARENA_PVE, &wg)
	wg.Wait()

	return int(elo), gift
}

func (HfUserArena) GiftEndgame(isWin bool) map[string]interface{} {
	ratioGift := 1
	giftCf := []interface{}{
		map[string]interface{}{constants.PERCENT: 11, constants.GIFT: map[string]interface{}{constants.STONES: map[string]int{constants.THUNDER: int(15 * ratioGift)}}},
		map[string]interface{}{constants.PERCENT: 11, constants.GIFT: map[string]interface{}{constants.STONES: map[string]int{constants.EARTH: int(15 * ratioGift)}}},
		map[string]interface{}{constants.PERCENT: 11, constants.GIFT: map[string]interface{}{constants.STONES: map[string]int{constants.FIRE: int(15 * ratioGift)}}},
		map[string]interface{}{constants.PERCENT: 11, constants.GIFT: map[string]interface{}{constants.STONES: map[string]int{constants.WATER: int(15 * ratioGift)}}},
		map[string]interface{}{constants.PERCENT: 11, constants.GIFT: map[string]interface{}{constants.STONES: map[string]int{constants.DARK: int(15 * ratioGift)}}},
		map[string]interface{}{constants.PERCENT: 11, constants.GIFT: map[string]interface{}{constants.STONES: map[string]int{constants.LIGHT: int(15 * ratioGift)}}},
		map[string]interface{}{constants.PERCENT: 12, constants.GIFT: map[string]interface{}{constants.STONES: map[string]int{constants.EVOLVE: int(10 * ratioGift)}}},
		map[string]interface{}{constants.PERCENT: 10, constants.GIFT: map[string]int{constants.GEM: int(5 * ratioGift)}},
	}
	if isWin {
		giftCf = append(giftCf, map[string]interface{}{constants.PERCENT: 12, constants.GIFT: map[string]int{constants.ARENA_COIN: int(20 * ratioGift)}})
	}
	gift := util.RandomPercentGift(giftCf)
	return gift
}

func (ua HfUserArena) GiftRank(db *gorm.DB, title string, serverId uint, typeLog, rank int, gifts iris.Map, group *sync.WaitGroup) {
	defer group.Done()

	uInbox := HfUserInbox{
		Id:          util.UUID(),
		ReceiverId:  ua.UserId,
		SenderType:  constants.INBOX_SENDER_BY_SYSTEM,
		ServerId:    serverId,
		TypeLog:     typeLog,
		KindId:      rank,
		Title:       title,
		Gift:        sql.NullString{String: util.JsonEndCode(gifts), Valid: true},
		CreatedDate: time.Now(),
	}

	go db.Create(&uInbox)
}
