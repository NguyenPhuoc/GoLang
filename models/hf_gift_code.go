package models

import (
	"GoLang/libraries/util"
	"database/sql"
	"encoding/json"
	"github.com/go-sql-driver/mysql"
	"sort"
	"time"
)

type HfGiftCode struct {
	Code      string `gorm:"primary_key"`
	Type      string
	Name      sql.NullString
	ServerId  string
	StartDate mysql.NullTime
	EndDate   mysql.NullTime
	Gift      string
	Quantity  int
	Used      int
	Note      sql.NullString
}

func (HfGiftCode) TableName() string {
	return "hf_gift_code"
}

func (gc HfGiftCode) CheckInput(u HfUser, checkInput bool) (bool, int64) {
	//check == true kiểm tra đầu zo
	//check == false nhập sai

	cacheValue := u.RedisInfo.HGet(u.UserId, gc.TableName())
	obj := struct {
		Quantity int       `json:"quantity"` //số lần nhập sai
		Time     time.Time `json:"time"`     //time kiểm tra nhập sai 5 lần or khóa 5p
	}{}

	if cacheValue.Err() == nil {
		_ = json.Unmarshal([]byte(cacheValue.Val()), &obj)
	}

	if checkInput {
		if obj.Quantity >= 10 && obj.Time.Add(time.Minute * 10).Unix() > time.Now().Unix() {
			return false, obj.Time.Add(time.Minute * 10).Unix() - time.Now().Unix()
		}
		return true, 0
	} else {
		if obj.Time.Add(time.Minute * 1).Unix() < time.Now().Unix() {
			obj.Quantity = 0
			obj.Time = time.Now()
		}

		obj.Quantity++
		if obj.Quantity >= 10 {
			obj.Time = time.Now()
		}

		u.RedisInfo.HSet(u.UserId, gc.TableName(), util.JsonEndCode(obj))
	}

	return true, 0
}

func (gc *HfGiftCode) Find(code string, u HfUser) (HfGiftCode, bool) {
	gc.Code = code
	field := gc.Code

	count := 0
	cacheValue := u.RedisConfig.HGet(gc.TableName(), field)
	if cacheValue.Err() != nil {
		u.DB.First(&gc).Count(&count)

		if count != 0 {
			u.RedisConfig.HSet(gc.TableName(), field, util.JsonEndCode(gc))
		}
	} else {
		count = 1
		_ = json.Unmarshal([]byte(cacheValue.Val()), &gc)
	}

	return *gc, count != 0
}

func (gc *HfGiftCode) GetAll(u HfUser) []HfGiftCode {
	results := []HfGiftCode{}

	check := u.RedisConfig.Exists(gc.TableName())
	if check.Val() == 0 {
		results = gc.CacheAll(u)
	} else {
		cacheValue := u.RedisConfig.HGetAll(gc.TableName())
		for _, val := range cacheValue.Val() {
			obj := HfGiftCode{}

			_ = json.Unmarshal([]byte(val), &obj)
			results = append(results, obj)
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Code < results[j].Code
	})

	return results
}

func (gc *HfGiftCode) CacheAll(u HfUser) []HfGiftCode {
	results := []HfGiftCode{}
	u.DB.Find(&results)

	u.RedisConfig.Del(gc.TableName())

	pipe := u.RedisConfig.Pipeline()
	for _, val := range results {
		pipe.HSet(gc.TableName(), val.Code, util.JsonEndCode(val))
	}
	_, _ = pipe.Exec()
	return results
}
