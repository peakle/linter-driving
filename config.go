package main

import (
	"github.com/vrischmann/envconfig"
)

type config struct {
	Token          string   `required:"true"`
	ProjectsDir    string   `envconfig:"default=projects"`
	LinterCloneURL string   `envconfig:"default=github.com/delivery-club/delivery-club-rules/tree/main/cmd/dcRules"`
	LinterArgs     []string `required:"true"`
}

func initConfig() (*config, error) {
	conf := &config{}
	if err := envconfig.Init(conf); err != nil {
		return nil, err
	}

	return conf, nil
}
