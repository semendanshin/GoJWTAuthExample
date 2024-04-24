package config

import (
	"flag"
	"github.com/ilyakaznacheev/cleanenv"
	"os"
	"time"
)

type Config struct {
	Env    string `yaml:"env" env-required:"true"` // dev, test, prod
	Tokens Tokens `yaml:"tokens"`
	Server Server `yaml:"server"`
	Mongo  Mongo  `yaml:"mongo"`
}

type Server struct {
	Address string        `yaml:"address" env-required:"true"`
	Timeout time.Duration `yaml:"timeout" env-required:"true"`
}

type Mongo struct {
	Host string `yaml:"host" env-required:"true"`
	Port int    `yaml:"port" env-required:"true"`
	User string `yaml:"user" env-required:"true"`
	Pass string `yaml:"pass" env-required:"true"`
	Name string `yaml:"name" env-required:"true"`
}

type Tokens struct {
	Secret     string        `yaml:"secret" env-required:"true"`
	AccessTTL  time.Duration `yaml:"access_ttl" env-required:"true"`
	RefreshTTL time.Duration `yaml:"refresh_ttl" env-required:"true"`
}

func MustParseConfig(path string) Config {
	var cfg Config

	err := cleanenv.ReadConfig(path, &cfg)
	if err != nil {
		panic(err)
	}

	return cfg
}

func FetchPath() string {
	var path string
	flag.StringVar(&path, "config", "", "path to config file")

	if path == "" {
		path = os.Getenv("CONFIG_PATH")
	}

	if path == "" {
		path = "config/local.yaml"
	}

	return path
}
