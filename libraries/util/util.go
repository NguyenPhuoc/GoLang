package util

import (
	"GoLang/libraries/constants"
	"GoLang/libraries/sessionmng"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/Knetic/govaluate"
	"github.com/google/uuid"
	"github.com/kataras/iris"
	"github.com/kataras/iris/sessions"
	"github.com/mintance/go-uniqid"
	"io"
	"math/rand"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

func InterfaceToMap(input interface{}) map[string]interface{} {
	v := reflect.ValueOf(input)
	output := iris.Map{}
	if v.Kind() == reflect.Map {
		for _, key := range v.MapKeys() {
			strct := v.MapIndex(key)
			output[fmt.Sprintf("%v", key.Interface())] = strct.Interface()
		}
	}

	return output
}

func InterfaceToArray(input interface{}) []interface{} {
	v := reflect.ValueOf(input)
	output := make([]interface{}, v.Len())
	if v.Kind() == reflect.Slice {
		for i := range output {
			output[i] = v.Index(i).Interface()
		}
	}

	return output
}

func ToArrayInt(input interface{}) []int {

	inputArr := InterfaceToArray(input)

	output := make([]int, len(inputArr))
	for i := range inputArr {
		output[i] = int(inputArr[i].(float64))
	}

	return output;
}

func UniqueInt(intSlice []int) []int {
	keys := make(map[int]bool)
	list := []int{}
	for _, entry := range intSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func JsonDecodeMap(jsonStr string) map[string]interface{} {
	jsonMap := make(map[string]interface{})

	_ = json.Unmarshal([]byte(jsonStr), &jsonMap)

	return jsonMap
}

func JsonDecodeArray(jsonStr string) []interface{} {
	jsonArray := make([]interface{}, 0)

	_ = json.Unmarshal([]byte(jsonStr), &jsonArray)

	return jsonArray
}

func JsonDecodeObject(jsonStr string, jsonObject interface{}) {

	_ = json.Unmarshal([]byte(jsonStr), &jsonObject)
}

func JsonDecodeProgress(jsonStr string) (map[int]int, []int) {
	jsonObject := map[int]int{}

	_ = json.Unmarshal([]byte(jsonStr), &jsonObject)

	// To store the keys in slice in sorted order
	var keys []int
	for k := range jsonObject {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	return jsonObject, keys
}

func JsonEndCode(object interface{}) string {
	endCode, _ := json.Marshal(object)

	return string(endCode)
}

//func MapKeys(mapInterface map[string]interface{}) []string {
func MapKeys(input interface{}) []string {

	//===============
	mapInterface := InterfaceToMap(input)

	keys := reflect.ValueOf(mapInterface).MapKeys()
	strkeys := make([]string, len(keys))

	for i := 0; i < len(keys); i++ {
		strkeys[i] = keys[i].String()
	}
	return strkeys
}

func MapKeysInt(input interface{}, sortOption int) []int {

	mapInterface := InterfaceToMap(input)

	keys := reflect.ValueOf(mapInterface).MapKeys()
	intKeys := make([]int, len(keys))

	for i := 0; i < len(keys); i++ {
		intKeys[i] = ToInt(keys[i].String())
	}

	if sortOption == 1 {
		sort.Slice(intKeys, func(i, j int) bool {
			return intKeys[i] < intKeys[j]
		})
	} else if sortOption == -1 {
		sort.Slice(intKeys, func(i, j int) bool {
			return intKeys[i] > intKeys[j]
		})
	}

	return intKeys
}

func ToString(i interface{}) string {
	return fmt.Sprintf("%v", i)
}

func ToInt(i interface{}) int {
	return ParseInt(ToString(i))
}

func ToInt64(i interface{}) int64 {
	num := int64(ToFloat(i))
	return num
}

func ToFloat(i interface{}) float64 {
	num, _ := strconv.ParseFloat(ToString(i), 64)
	return num
}

func ParseInt(str string) int {
	num, _ := strconv.ParseFloat(str, 64)
	return int(num)
}

func Time() time.Time {
	return time.Unix(time.Now().Unix(), 0)
}

func TimeUnix() int64 {
	return time.Now().Unix()
}

func GetTokenValid(sess *sessions.Session) map[string]int64 {

	baseTokenKey := regexp.MustCompile("token_[a-z0-9]{32}$")
	sessAll := sess.GetAll()

	tokenFormat := make(map[string]int64)

	for k, v := range sessAll {
		tokenKey := baseTokenKey.FindString(k)
		if len(tokenKey) == 38 {
			expired := int64(v.(float64))

			if TimeUnix() <= expired {
				token := strings.Replace(tokenKey, "token_", "", -1)
				tokenFormat[token] = expired
			} else {
				sess.Delete(tokenKey)
			}
		}
	}

	return tokenFormat
}

func GetToken(sess *sessions.Session) string {
	uniq := uniqid.New(uniqid.Params{"", true})
	token := fmt.Sprintf("%x", md5.Sum([]byte(uniq)))

	//GetTokenValid(sess)

	tokenKey := fmt.Sprintf("%s_%s", sessionmng.KEY_TOKEN, token)
	tokenExpires := TimeUnix() + sessionmng.EXPIRES_TOKEN
	//sess.SetFlash(tokenKey, tokenExpires)
	sess.Set(tokenKey, tokenExpires)

	return token
}

func ValidToken(token string, sign string, sess *sessions.Session) (bool, interface{}) {
	hash := md5.New()
	_, _ = io.WriteString(hash, constants.HASH_TOKEN)
	_, _ = io.WriteString(hash, token)
	signHash := fmt.Sprintf("%x", hash.Sum(nil))

	//tokensMap := GetTokenValid(sess)

	tokenKey := fmt.Sprintf("%s_%s", sessionmng.KEY_TOKEN, token)
	//tokenExpires, _ := sess.GetFlash(tokenKey).(int64)
	tokenExpires, err := sess.GetInt64(tokenKey)
	sess.Delete(tokenKey)

	msg := iris.Map{}
	msg["key"] = tokenKey
	msg["tokenExpires"] = tokenExpires
	msg["err"] = err
	msg["time_check"] = tokenExpires - TimeUnix()
	msg["sign_check"] = sign == signHash

	return tokenExpires >= TimeUnix() && sign == signHash, msg
}

func ValidHash(form func(string) string, data ...interface{}) bool {

	hash := form("hash")

	hashMd5 := md5.New()
	for _, val := range data {
		_, _ = io.WriteString(hashMd5, ToString(val))
	}

	_, _ = io.WriteString(hashMd5, form("sign"))
	signHash := fmt.Sprintf("%x", hashMd5.Sum(nil))

	return hash == signHash
}

func UUID() string {
	newUUID, _ := uuid.NewRandom()
	return newUUID.String()
}

func QuantityUint(quantity interface{}, quantityMore int) uint {

	newQuantity := ToInt(quantity);
	newQuantity += quantityMore;
	if newQuantity < 0 {
		newQuantity = 0
	}

	return uint(newQuantity)
}

func InArray(val interface{}, array interface{}) (exists bool) {
	exists = false
	//index = -1

	switch reflect.TypeOf(array).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(array)

		for i := 0; i < s.Len(); i++ {
			if reflect.DeepEqual(val, s.Index(i).Interface()) == true {
				//index = i
				exists = true
				return
			}
		}
	}

	return
}

func RandomPercentGift(gifts interface{}, ran ...int) map[string]interface{} {
	giftsArr := InterfaceToArray(gifts)

	random := 0
	if len(ran) == 0 {
		totalRatio := 0

		for _, item := range giftsArr {
			item := InterfaceToMap(item)
			if per, ok := item[constants.PERCENT]; ok {
				totalRatio += ToInt(per)
			}
		}
		random = rand.Intn(totalRatio)

	} else {
		random = ran[0]
	}

	percent := 0
	for _, item := range giftsArr {
		item := InterfaceToMap(item)
		if per, ok := item[constants.PERCENT]; ok {
			percent += ToInt(per)
			if percent > random {
				if gift, ok := item[constants.GIFT]; ok {
					return InterfaceToMap(gift)
				} else if giftRand, ok := item[constants.RAND]; ok {
					return RandGift(giftRand)
				}
			}
		}
	}

	return InterfaceToMap(nil)
}

//Gift, index, id
func RandomPercentGiftEvent(gifts interface{}, ran ...int) (giftReturn map[string]interface{}, indexGift int, idGift int) {
	giftReturn, indexGift, idGift = map[string]interface{}{}, 0, 0

	giftsArr := InterfaceToArray(gifts)

	random := 0
	if len(ran) == 0 {
		totalRatio := 0

		for _, item := range giftsArr {
			item := InterfaceToMap(item)
			if per, ok := item[constants.PERCENT]; ok {
				totalRatio += ToInt(per)
			}
		}
		random = rand.Intn(totalRatio)

	} else {
		random = ran[0]
	}

	percent := 0
	for _, item := range giftsArr {
		item := InterfaceToMap(item)
		if per, ok := item[constants.PERCENT]; ok {
			percent += ToInt(per)
			if percent > random {

				if val, ok := item[constants.ID]; ok {
					idGift = ToInt(val)
				}
				if val, ok := item[constants.INDEX]; ok {
					indexGift = ToInt(val)
				}

				if gift, ok := item[constants.GIFT]; ok {
					giftReturn = InterfaceToMap(gift)
					return
				} else if giftRand, ok := item[constants.RAND]; ok {
					giftReturn = RandGift(giftRand)
					return
				}
			}
		}
	}

	return
}

func RandGift(giftRand interface{}) map[string]interface{} {
	giftMap := InterfaceToMap(giftRand)

	//{"rand":{"stones":{"all":3}},"percent":40}
	if gift, ok := giftMap[constants.STONES]; ok {
		gift := InterfaceToMap(gift)
		if quantity, ok := gift[constants.ALL]; ok {
			quantity := ToInt(quantity)
			allStone := []string{constants.FIRE, constants.EARTH, constants.THUNDER, constants.WATER, constants.LIGHT, constants.DARK}
			stoneType := allStone[rand.Intn(len(allStone))]

			return map[string]interface{}{constants.STONES: map[string]int{stoneType: quantity}}
		}
	}

	//{"rand":{"equip":{"level":6,"back":2,"quantity":1}},"percent":20}
	if gift, ok := giftMap[constants.EQUIP]; ok {
		gift := InterfaceToMap(gift)
		level := uint16(ToInt(gift[constants.LEVEL]))
		back := uint16(ToInt(gift[constants.BACk]))
		quantity := ToInt(gift[constants.QUANTITY])

		allEquips := []uint16{
			101, 201, 301, 401, 501, 601,
			102, 202, 302, 402, 502, 602,
			103, 203, 303, 403, 503, 603,
			104, 204, 304, 404, 504, 604,
			105, 205, 305, 405, 505, 605,
			106, 206, 306, 406, 506, 606,
			107, 207, 307, 407, 507, 607,
			108, 208, 308, 408, 508, 608,
			109, 209, 309, 409, 509, 609,
			110, 210, 310, 410, 510, 610,
			111, 211, 311, 411, 511, 611,
			112, 212, 312, 412, 512, 612,
			113, 213, 313, 413, 513, 613,
			114, 214, 314, 414, 514, 614,
			115, 215, 315, 415, 515, 615,
			116, 216, 316, 416, 516, 616,
			117, 217, 317, 417, 517, 617,
			118, 218, 318, 418, 518, 618,
			119, 219, 319, 419, 519, 619,
			120, 220, 320, 420, 520, 620,
		}
		idsEquipRan := []uint16{}
		for _, id := range allEquips {
			if id%100 > level-back && id%100 <= level {
				idsEquipRan = append(idsEquipRan, id)
			}
		}

		equipsId := idsEquipRan[rand.Intn(len(idsEquipRan))]

		return map[string]interface{}{constants.EQUIP: map[uint16]int{equipsId: quantity}}
	}

	//{"rand":{"piece_general":{"type":1,"back":1,"branch":"g","quantity":1}},"percent":40}
	if gift, ok := giftMap[constants.PIECE_GENERAL]; ok {
		gift := InterfaceToMap(gift)
		typePiece := ToInt(gift[constants.TYPE])
		typeBack := ToInt(gift[constants.BACk])
		branchPiece := ToString(gift[constants.BRANCH])
		quantity := ToInt(gift[constants.QUANTITY])

		max := typePiece
		min := typePiece - typeBack + 1
		typeRand := rand.Intn(max-min+1) + min

		return map[string]interface{}{constants.PIECE_GENERAL: []map[string]interface{}{
			{
				constants.BRANCH:   branchPiece,
				constants.TYPE:     typeRand,
				constants.QUANTITY: quantity,
			},
		}}
	}

	return InterfaceToMap(nil)
}

func MergeGift(gifts interface{}, gift interface{}) map[string]interface{} {

	giftsMap := InterfaceToMap(gifts)
	giftMap := InterfaceToMap(gift)

	for key, valGift := range giftMap {

		if valGifts, ok := giftsMap[key]; ok {
			switch key {
			case constants.GEM, constants.GOLD, constants.EXP, constants.TICKET_ARENA, constants.POWER_TOWER, constants.TICKET_ARENA_PVP,
				constants.TICKET_GUARDIAN, constants.TICKET_MARKET_BLACK, constants.FLOWER_GUARDIAN, constants.FRUIT_GUARDIAN, constants.STONE_GUARDIAN, constants.GEM_SOUL:
				giftsMap[key] = ToInt(valGifts) + ToInt(valGift)

			case constants.EQUIP, constants.STONES, constants.PIECE, constants.SUMMON_BALL, constants.PIECE_GUARDIAN:
				giftItem := InterfaceToMap(valGift)
				giftItems := InterfaceToMap(valGifts)

				for id, quanGift := range giftItem {
					if quantity, ok := giftItems[id]; ok {
						giftItems[id] = ToInt(quantity) + ToInt(quanGift)
					} else {
						giftItems[id] = giftItem[id]
					}
				}
				giftsMap[key] = giftItems

			case constants.PIECE_GENERAL:
				giftItem := InterfaceToArray(valGift)
				giftItems := InterfaceToArray(valGifts)

				for _, valG := range giftItem {
					valG := InterfaceToMap(valG)

					checkG := false

					for i, valGs := range giftItems {
						valGs := InterfaceToMap(valGs)

						if valG["type"] == valGs["type"] && valG["branch"] == valGs["branch"] {

							valGs["quantity"] = ToInt(valGs["quantity"]) + ToInt(valG["quantity"])
							giftItems[i] = valGs

							checkG = true
							break
						}
					}

					if !checkG {
						giftItems = append(giftItems, valG)
					}
					giftsMap[key] = giftItems
				}

			}

		} else {
			giftsMap[key] = valGift
		}
	}

	return giftsMap
}

func TimeToDate(t time.Time) string {

	timeDate := fmt.Sprintf("%d-%02d-%02d", t.Year(), t.Month(), t.Day())

	return timeDate
}

func TimeToDateTime(t time.Time) string {

	timeDate := fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())

	return timeDate
}

func TimeHHiissddmmYYYY(t time.Time) string {

	timeDate := fmt.Sprintf("%02d:%02d:%02d %02d-%02d-%d", t.Hour(), t.Minute(), t.Second(), t.Day(), t.Month(), t.Year())

	return timeDate
}

func StatusGift(status interface{}) string {
	switch ToInt(status) {
	case 0:
		return "claim" //là bấm nút nhận đi
	case 1:
		return "done" //là đã nhận rồi, xong hết rồi
	case 2:
		return "none" //là chưa có gì hết
	default:
		return "none"
	}
}

func RanCode(length int) string {
	var letterRunes = []rune("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ")

	code := make([]rune, length)
	for i := range code {
		code[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(code)
}

func EvaluateMath(expr string, parameters map[string]interface{}) interface{} {

	expr = strings.Replace(expr, "^", "**", -1)

	expression, err := govaluate.NewEvaluableExpression(expr)
	if err != nil {
		fmt.Println(err)
	}

	result, err := expression.Evaluate(parameters)
	if err != nil {
		fmt.Println(err)
	}

	return result
}
