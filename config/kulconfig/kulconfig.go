package kulconfig

import (
	"GoLang/libraries/util"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/tkanos/gonfig"
	"time"
)

type kulconfig struct {
	AppKey           string
	SecretKey        string
	SecretKeyPayment string
	API              string

	AccessToken string
	Time        string
	Sign        string
}

func Load(accessToken string) kulconfig {
	kul := kulconfig{}
	err := gonfig.GetConf("config/kulconfig/kulconfig.json", &kul)
	if err != nil {
		fmt.Println(err.Error())
	}

	t := time.Now()
	timeFormat := fmt.Sprintf("%d%02d%02d%02d%02d%02d",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second())

	h := hmac.New(sha256.New, []byte(kul.SecretKey))
	h.Write([]byte(accessToken))
	h.Write([]byte(timeFormat))
	sign := hex.EncodeToString(h.Sum(nil))

	kul.AccessToken = accessToken
	kul.Time = timeFormat
	kul.Sign = sign

	return kul
}

func CheckSign(sign string, data ...interface{}) (bool, string) {
	kul := kulconfig{}
	err := gonfig.GetConf("config/kulconfig/kulconfig.json", &kul)
	if err != nil {
		fmt.Println(err.Error())
		return false, err.Error()
	}

	h := hmac.New(sha256.New, []byte(kul.SecretKeyPayment))
	for _, val := range data {
		h.Write([]byte(util.ToString(val)))
	}
	hash := hex.EncodeToString(h.Sum(nil))

	if sign != hash {
		return false, `Sign invalid`
	}

	return true, `Valid`
}
