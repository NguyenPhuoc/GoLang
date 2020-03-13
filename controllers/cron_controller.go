package controllers

import (
	"GoLang/libraries/logtype"
	"GoLang/libraries/util"
	"GoLang/models"
	"fmt"
	"github.com/kataras/iris"
	"sync"
)

type CronController struct {
	MyController
}

// /cron/demo
func (c *CronController) AnyDemo(form formValue, ctx iris.Context) {

	serverId := uint(util.ToInt(form("server_id")))

	fmt.Println("server_id", serverId)
	fmt.Println("DONE")
}

// /cron/arena/daily
func (c *CronController) AnyArenaDaily(form formValue, ctx iris.Context) {

	serverId := uint(util.ToInt(form("server_id")))

	tops := []struct {
		UserId   string `json:"user_id"`
		FullName string `json:"full_name"`
		Elo      int    `json:"elo"`
		Top      int    `json:"top"`
	}{}
	c.DB.Raw(`SELECT u.user_id, u.full_name, elo FROM hf_user_arena are 
				JOIN hf_user u ON u.user_id = are.user_id
				WHERE is_boss = 0 AND hit_daily > 0 AND server_id = ?
				ORDER BY elo DESC;`, serverId).Scan(&tops)

	arenaAward := models.HfArenaAward{}
	gift_cf := arenaAward.GetAll(c.User)

	var wg sync.WaitGroup
	for i, val := range tops {
		val.Top = i + 1

		gifts := iris.Map{}
		for _, aAward := range gift_cf {
			if val.Top >= aAward.Rank {
				gifts = util.JsonDecodeMap(aAward.BoxDay)

				wg.Add(1)
				uArena := models.HfUserArena{UserId: val.UserId}
				go uArena.GiftRank(c.DB, "Quà xếp hạng đấu trường Vinh Quang ngày", serverId, logtype.GIFT_ARENA_RANK_DAILY, val.Top, gifts, &wg)

				break
			}
		}
	}
	wg.Wait()

	c.DB.Exec("UPDATE hf_user_arena are JOIN hf_user u ON u.user_id = are.user_id SET are.hit_daily = 0 WHERE is_boss = 0 AND hit_daily > 0 AND server_id = ?;", serverId)

	c.DataResponse = iris.Map{"code": 1, "msg": "GIFT_ARENA_RANK_DAILY", "server_id": serverId}
}

// /cron/arena/weekly
func (c *CronController) AnyArenaWeekly(form formValue, ctx iris.Context) {

	serverId := uint(util.ToInt(form("server_id")))

	tops := []struct {
		UserId   string `json:"user_id"`
		FullName string `json:"full_name"`
		Elo      int    `json:"elo"`
		Top      int    `json:"top"`
	}{}
	c.DB.Raw(`SELECT u.user_id, u.full_name, elo FROM hf_user_arena are 
				JOIN hf_user u ON u.user_id = are.user_id
				WHERE is_boss = 0 AND hit_weekly > 0 AND server_id = ?
				ORDER BY elo DESC;`, serverId).Scan(&tops)

	arenaAward := models.HfArenaAward{}
	gift_cf := arenaAward.GetAll(c.User)

	var wg sync.WaitGroup
	for i, val := range tops {
		val.Top = i + 1

		gifts := iris.Map{}
		for _, aAward := range gift_cf {
			if val.Top >= aAward.Rank {
				gifts = util.JsonDecodeMap(aAward.BoxWeek)

				wg.Add(1)
				uArena := models.HfUserArena{UserId: val.UserId}
				go uArena.GiftRank(c.DB, "Quà xếp hạng đấu trường Vinh Quang tuần", serverId, logtype.GIFT_ARENA_RANK_WEEKLY, val.Top, gifts, &wg)

				break
			}
		}
	}
	wg.Wait()

	c.DB.Exec("UPDATE hf_user_arena are JOIN hf_user u ON u.user_id = are.user_id SET are.hit_weekly = 0, elo = IF(elo < 1000, 1000, (elo-1000)/2 + 1000) WHERE is_boss = 0 AND hit_weekly > 0 AND server_id = ?;", serverId)
	//Todo IF(elo < 1000, 1000, (elo-1000)/2 + 1000)

	c.DataResponse = iris.Map{"code": 1, "msg": "GIFT_ARENA_RANK_WEEKLY", "server_id": serverId}
}

// /cron/arena/pvp/weekly
func (c *CronController) AnyArenaPvpWeekly(form formValue, ctx iris.Context) {

	serverId := uint(util.ToInt(form("server_id")))

	tops := []struct {
		UserId   string `json:"user_id"`
		FullName string `json:"full_name"`
		Elo      int    `json:"elo"`
		Top      int    `json:"top"`
	}{}
	c.DB.Raw(`SELECT u.user_id, u.full_name, elo FROM hf_user_arena_pvp are 
				JOIN hf_user u ON u.user_id = are.user_id
				WHERE hit_weekly > 0 AND server_id = ?
				ORDER BY elo DESC;`, serverId).Scan(&tops)

	pvpAward := models.HfArenaPvpAward{}
	gift_cf := pvpAward.GetAll(c.User)

	var wg sync.WaitGroup
	for i, val := range tops {
		val.Top = i + 1

		gifts := iris.Map{}
		for _, pAward := range gift_cf {
			if val.Top >= pAward.Rank {
				gifts = util.JsonDecodeMap(pAward.BoxWeek)

				wg.Add(1)
				uArena := models.HfUserArena{UserId: val.UserId}
				go uArena.GiftRank(c.DB, "Quà xếp hạng đấu trường Thách Đấu tuần", serverId, logtype.GIFT_ARENA_PVP_RANK_WEEKLY, val.Top, gifts, &wg)

				break
			}
		}
	}
	wg.Wait()

	c.DB.Exec("UPDATE hf_user_arena_pvp are JOIN hf_user u ON u.user_id = are.user_id SET are.hit_weekly = 0, elo = IF(elo < 1000, 1000, (elo-1000)/2 + 1000) WHERE hit_weekly > 0 AND server_id = ?;", serverId)

	c.DataResponse = iris.Map{"code": 1, "msg": "GIFT_ARENA_PVP_RANK_WEEKLY", "server_id": serverId}
}
