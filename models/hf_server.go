package models

import (
	"GoLang/config/configdb"
	"GoLang/libraries/util"
	"database/sql"
	"encoding/json"
	"github.com/kataras/iris"
	"sort"
	"time"
)

type HfServer struct {
	ServerId uint `gorm:"primary_key"`
	Name     sql.NullString
	DbName   string
	New      uint8
	DateOpen time.Time
}

func (HfServer) TableName() string {
	return "hf_server"
}

func (s *HfServer) GetMap() iris.Map {
	return iris.Map{
		"server_id": s.ServerId,
		"name":      s.Name.String,
		"db_name":   s.DbName,
		"date_open": s.DateOpen.String(),
	}
}

func (s *HfServer) Find(serverId uint, u HfUser) (HfServer, bool) {
	s.ServerId = serverId
	field := util.ToString(serverId)

	count := 0
	cacheValue := u.RedisConfig.HGet(s.TableName(), field)
	if cacheValue.Err() != nil {
		u.DB.Where("server_id = ?", s.ServerId).First(&s).Count(&count)
		s.CacheAll(u)
	} else {
		count = 1
		_ = json.Unmarshal([]byte(cacheValue.Val()), &s)
	}

	return *s, count != 0
}

func (s *HfServer) GetDBName(serverId uint, u HfUser) string {
	sv, check := s.Find(serverId, u)

	if !check {
		sv.DbName = configdb.MAIN_DB
	}

	return sv.DbName
}

func (s *HfServer) GetDBs(u HfUser) []uint {
	allServer := s.GetAll(u)

	svString := []string{}
	svUint := []uint{}
	for _, sv := range allServer {
		if !util.InArray(sv.DbName, svString) {
			svString = append(svString, sv.DbName)
			svUint = append(svUint, sv.ServerId)
		}
	}

	return svUint
}

func (s *HfServer) GetAll(u HfUser) []HfServer {
	results := []HfServer{}

	check := u.RedisConfig.Exists(s.TableName())
	if check.Val() == 0 {
		results = s.CacheAll(u)
	} else {
		cacheValue := u.RedisConfig.HGetAll(s.TableName())
		for _, val := range cacheValue.Val() {
			obj := HfServer{}

			_ = json.Unmarshal([]byte(val), &obj)
			results = append(results, obj)
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].ServerId < results[j].ServerId
	})

	return results
}

func (s *HfServer) CacheAll(u HfUser) []HfServer {
	results := []HfServer{}
	u.DB.Find(&results)

	u.RedisConfig.Del(s.TableName())

	pipe := u.RedisConfig.Pipeline()
	for _, val := range results {
		pipe.HSet(s.TableName(), util.ToString(val.ServerId), util.JsonEndCode(val))
	}
	_, _ = pipe.Exec()
	return results
}
