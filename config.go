package main

import (
	"github.com/vrischmann/envconfig"
)

type config struct {
	Token            string   `required:"true"`
	ExcludedProjects []string `envconfig:"default=kubernetes;the-way-to-go_ZH_CN"`
	ProjectsDir      string   `envconfig:"default=projects"`
	LinterCloneURL   string   `envconfig:"default=https://github.com/delivery-club/delivery-club-rules"`
	PathToMain       string   `envconfig:"default=/cmd/dcRules"`
	BinaryName       string   `envconfig:"default=dcRules"`
	LinterArgs       []string `required:"true"`
}

func initConfig() (*config, error) {
	conf := &config{}
	err := envconfig.Init(conf)

	return conf, err
}
