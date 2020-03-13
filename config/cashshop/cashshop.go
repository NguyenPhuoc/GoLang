package cashshop

import (
	"fmt"
	"github.com/tkanos/gonfig"
)

type cashshop struct {
	PackageId   string
	PackageName string
	Web         int
	Google      int
	Ios         int
	Gem         int
	GemFirst    int
}

func Config() map[string]cashshop {
	conf := struct {
		Config map[string]cashshop
	}{}
	err := gonfig.GetConf("config/cashshop/cashshop.json", &conf)
	if err != nil {
		fmt.Println(err.Error())
	}

	return conf.Config
}
