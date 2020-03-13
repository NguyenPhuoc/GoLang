package environment

import (
	"github.com/tkanos/gonfig"
	"sync"
)

const (
	LOCAL      = "LOCAL"
	DEV        = "DEV"
	PRODUCTION = "PRODUCTION"
)

type environment struct {
	ENVIRONMENT string
	Addr        string
}

var env *environment
var once sync.Once

func LoadEnvironmentLocal() *environment {
	once.Do(func() {
		envz := environment{}
		err := gonfig.GetConf("config/environment/environment.json", &envz)
		if err != nil {
			panic(err)
		}
		env = &envz
	})
	return env
}

func LoadEnvironment() environment {
	env := environment{}
	err := gonfig.GetConf("config/environment/environment.json", &env)
	if err != nil {
		panic(err)
	}
	return env
}
