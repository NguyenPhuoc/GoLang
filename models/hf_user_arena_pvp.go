package models

import (
	"GoLang/libraries/constants"
	"GoLang/libraries/logtype"
	"GoLang/libraries/util"
	"database/sql"
	"github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/kataras/iris"
	"math"
	"math/rand"
	"sync"
	"time"
)

func (HfUserArenaPvp) GetPriceTicket() int {
	//Giá 3 ticket là 100 gem
	return 100
}
func (HfUserArenaPvp) GetDayTicket() uint {
	//Mỗi ngày nhận 3 vé đấu trường(3 tim)
	return 3
}

type HfUserArenaPvp struct {
	UserId      string `gorm:"primary_key"`
	FullName    string `gorm:"default:null"`
	Ticket      uint   `gorm:"default:3"`
	Elo         int    `gorm:"default:1000"`
	PowerPoint  int
	Level       int
	LastWin     int `gorm:"default:-1"`
	RivalElo    int
	Battle      uint
	TicketDate  mysql.NullTime
	RecentRival sql.NullString
	LastLineUp  sql.NullString
	HitWeekly   int
}

func (HfUserArenaPvp) TableName() string {
	return "hf_user_arena_pvp"
}

func (ua *HfUserArenaPvp) GetMap() iris.Map {
	return iris.Map{
		"user_id":                  ua.UserId,
		"full_name":                ua.FullName,
		constants.TICKET_ARENA_PVP: ua.Ticket,
		"elo_pvp":                  ua.Elo,
		"power_point":              ua.PowerPoint,
		"recent_rival":             ua.RecentRival.String,
	}
}

func (ua *HfUserArenaPvp) Get(u HfUser) HfUserArenaPvp {
	ua.UserId = u.UserId

	count := 0
	u.DB.First(&ua).Count(&count)

	if count == 0 {
		ua.FullName = u.FullName
		ua.Elo = 1000
		ua.Ticket = 3
		ua.TicketDate = mysql.NullTime{Time: time.Now(), Valid: true}
		u.DB.Create(&ua)
	} else if util.TimeToDate(time.Now()) != util.TimeToDate(ua.TicketDate.Time) {
		ua.TicketDate = mysql.NullTime{Time: time.Now(), Valid: true}
		ua.Battle = 0

		if ua.Ticket < ua.GetDayTicket() {
			var wg sync.WaitGroup
			wg.Add(1)

			ticket := int(ua.GetDayTicket() - ua.Ticket)
			go ua.SetTicket(ticket, "", nil, logtype.GET_DAY_TICKET_ARENA_PVP, 0, u, &wg)
			wg.Wait()
		} else {
			u.DB.Save(&ua)
		}
	}

	return *ua
}

func (ua *HfUserArenaPvp) Find(userId string, db *gorm.DB) (HfUserArenaPvp, bool) {
	ua.UserId = userId

	count := 0
	if userId != "" {
		db.First(&ua).Count(&count)
	}

	return *ua, count != 0
}

func (ua *HfUserArenaPvp) SetTicket(quantity int, itemId interface{}, kindId interface{}, typeLog, eventId int, u HfUser, wg *sync.WaitGroup) {
	defer wg.Done()

	ua.Ticket = util.QuantityUint(ua.Ticket, quantity)

	wg.Add(2)
	go u.SaveLog(typeLog, constants.TICKET_ARENA_PVP, itemId, kindId, eventId, quantity, uint64(ua.Ticket), "", wg)

	go func() {
		u.DB.Save(&ua)
		wg.Done()
	}()
}

func (ua *HfUserArenaPvp) GetRankAndTop(db *gorm.DB) (int, string) {

	tops := []struct {
		UserId     string `json:"user_id"`
		FullName   string `json:"full_name"`
		Elo        int    `json:"elo_pvp"`
		Level      int    `json:"level"`
		PowerPoint int    `json:"power_point"`
		Top        int    `json:"top"`
		AvatarId   int    `json:"avatar_id"`
	}{}
	//db.Raw(`SELECT user_id, full_name, elo, level, power_point FROM hf_user_arena_pvp ORDER BY elo DESC LIMIT 100;`).Scan(&tops)
	//db.Raw(`SELECT are.user_id, full_name, elo, level, power_point, ava.avatar_id FROM hf_user_arena_pvp are
	//			LEFT JOIN (SELECT user_id, MAX(avatar_id) avatar_id FROM hf_user_avatar WHERE used = 1 GROUP BY user_id) ava
	//			ON ava.user_id = are.user_id
	//			WHERE power_point != 0
	//			ORDER BY elo DESC LIMIT 100;`).Scan(&tops)
	db.Raw(`SELECT are.user_id, are.full_name, elo, level, power_point, ava.avatar_id FROM hf_user_arena_pvp are 
				LEFT JOIN hf_user ava 
				ON ava.user_id = are.user_id 
				WHERE power_point != 0 
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

func (ua *HfUserArenaPvp) EndgameWinLose(isWin bool, data string, u HfUser) (int, interface{}) {

	elo := float64(0);
	if isWin {
		//Todo Elo cộng =MIN(45;MAX(10;30 - (elo bản thân - elo đối thủ)/100) )
		elo = math.Min(45, math.Max(1, float64(30-(ua.Elo-ua.RivalElo)/100)))
	} else {
		//Todo Elo thua = MIN(-10;MAX(-45;(elo đối thủ - elo bản thân)/100 -30))*70% )
		elo = math.Min(-10, math.Max(-45, float64((ua.RivalElo-ua.Elo)/100-30))*0.7)
	}
	if ua.Battle >= 15 {
		elo = 0
	}
	ua.Battle++
	ua.Elo += int(elo)

	//save log
	var wg sync.WaitGroup
	ual := HfUserArenaPvpLog{UserId: ua.UserId, Elo: int(elo), IsWin: ua.LastWin, EloAfter: ua.Elo, Data: data}
	//go u.DB.Create(&ual)
	wg.Add(1)
	go u.SaveLogMongo(ual.TableName(), iris.Map{"server_id": u.ServerId, "user_id": ua.UserId, "elo": int(elo), "is_win": ua.LastWin, "elo_after": ua.Elo, "data": util.JsonDecodeMap(data), "created_date": time.Now()}, &wg)

	ua.LastWin = -1
	ua.RivalElo = 0
	u.DB.Save(&ua)

	//update gift
	wg.Add(2)
	gift := ua.GiftEndgame(isWin, ua.Battle, ua.Elo)
	go u.UpdateGifts(gift, logtype.GIFT_ENDGAME_ARENA_PVP, 0, &wg)
	//Todo Update nhiệm vụ
	go u.CompleteMissionDaily(QDL_ARENA_PVP, &wg)
	wg.Wait()

	return int(elo), gift
}

func (HfUserArenaPvp) GiftEndgame(isWin bool, battle uint, elo int) map[string]interface{} {
	ratioGift := 1 + float64(elo) / 1000

	giftWin := []iris.Map{
		{constants.INDEX: 1, constants.PERCENT: 1, constants.GIFT: iris.Map{constants.STONES: map[string]int{constants.EVOLVE: int(5 * ratioGift)}}},
		{constants.INDEX: 2, constants.PERCENT: 1, constants.GIFT: iris.Map{constants.GEM_SOUL: int(2 * ratioGift)}},
		{constants.INDEX: 3, constants.PERCENT: 1, constants.GIFT: iris.Map{constants.GEM: int(2 * ratioGift)}},
		{constants.INDEX: 4, constants.PERCENT: 1, constants.GIFT: iris.Map{constants.ARENA_COIN: int(4 * ratioGift)}},
	}

	giftLose := []iris.Map{
		{constants.PERCENT: 1, constants.GIFT: map[string]interface{}{constants.STONES: map[string]int{constants.THUNDER: int(5 * ratioGift)}}},
		{constants.PERCENT: 1, constants.GIFT: map[string]interface{}{constants.STONES: map[string]int{constants.EARTH: int(5 * ratioGift)}}},
		{constants.PERCENT: 1, constants.GIFT: map[string]interface{}{constants.STONES: map[string]int{constants.FIRE: int(5 * ratioGift)}}},
		{constants.PERCENT: 1, constants.GIFT: map[string]interface{}{constants.STONES: map[string]int{constants.WATER: int(5 * ratioGift)}}},
		{constants.PERCENT: 1, constants.GIFT: map[string]interface{}{constants.STONES: map[string]int{constants.DARK: int(5 * ratioGift)}}},
		{constants.PERCENT: 1, constants.GIFT: map[string]interface{}{constants.STONES: map[string]int{constants.LIGHT: int(5 * ratioGift)}}},
	}

	quantityGift := 1
	ixGifts := []int{}
	gifts := iris.Map{}

	if battle <= 10 && isWin {
		quantityGift = rand.Intn(3-2) + 2
	}

	if battle <= 10 || isWin {
		for len(ixGifts) < quantityGift {
			gift, inx, _ := util.RandomPercentGiftEvent(giftWin)
			if !util.InArray(inx, ixGifts) {
				ixGifts = append(ixGifts, inx)
				gifts = util.MergeGift(gifts, gift)
			}
		}
	} else {
		gifts = util.RandomPercentGift(giftLose)
	}

	return gifts
}
