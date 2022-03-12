package main

import (
	"github.com/vrischmann/envconfig"
)

type Config struct {
	Token          string   `required:"true"`
	ProjectsDir    string   `required:"true"`
	LinterCloneURL string   `require:"true"`
	LinterArgs     []string `required:"true"`
}

func InitConfig() (*Config, error) {
	conf := &Config{}
	if err := envconfig.Init(conf); err != nil {
		return nil, err
	}

	return conf, nil
}
