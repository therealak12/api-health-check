package config

import (
	"time"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/structs"
	"github.com/sirupsen/logrus"
)

type (
	Config struct {
		Logger   Logger   `koanf:"logger"`
		Database Database `koanf:"database"`
		Webhook  Webhook  `koanf:"webhook"`
	}

	Logger struct {
		Level   logrus.Level `koanf:"level"`
		Enabled bool         `koanf:"enabled"`
	}

	Database struct {
		Host     string        `koanf:"host"`
		Port     string        `koanf:"port"`
		Name     string        `koanf:"name"`
		Username string        `koanf:"username"`
		Password string        `koanf:"password"`
		Timeout  time.Duration `koanf:"timeout"`
	}

	Webhook struct {
		Url              string        `koanf:"url"`
		MessageFieldName string        `koanf:"messageFieldName"`
		Timeout          time.Duration `koanf:"timeout"`
	}
)

var defaultConfig = Config{
	Logger: Logger{
		Level:   5,
		Enabled: true,
	},
	Database: Database{
		Host:     "localhost",
		Port:     "5432",
		Name:     "healthcheck",
		Username: "healthcheck",
		Password: "healthcheck",
		Timeout:  5 * time.Second,
	},
	Webhook: Webhook{
		Url:              "http://localhost:5050",
		MessageFieldName: "message",
		Timeout:          5,
	},
}

func New() Config {
	var instance Config

	k := koanf.New(".")

	if err := k.Load(structs.Provider(defaultConfig, "koanf"), nil); err != nil {
		logrus.Fatalf("error loading default: %s", err)
	}

	if err := k.Load(file.Provider("config.yaml"), yaml.Parser()); err != nil {
		logrus.Errorf("error loading file: %s", err)
	}

	if err := k.Unmarshal("", &instance); err != nil {
		logrus.Fatalf("error unmarshalling config: %s", err)
	}

	return instance
}
