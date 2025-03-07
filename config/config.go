package config

import (
	"log"
	"os"

	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v2"
)

type Config struct {
	MongoDB struct {
		URI     string `yaml:"uri"`
		UserDB  string `yaml:"user_db"`  // 用户数据存储的数据库
		OrderDB string `yaml:"order_db"` // 订单数据存储的数据库
	} `yaml:"mongodb"`
	Redis struct {
		Host     string `yaml:"host"`
		Port     string `yaml:"port"`
		Password string `yaml:"password"`
		UserDB   int    `yaml:"user_db"`  // 用户数据缓存库
		OrderDB  int    `yaml:"order_db"` // 订单数据缓存库
	} `yaml:"redis"`
	Server struct {
		Port string `yaml:"port"`
	} `yaml:"server"`
	LogLevel string `yaml:"log_level"`
}

var AppConfig Config

func GetLogLevel() zapcore.Level {
	switch AppConfig.LogLevel {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

func InitConfig() {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "dev"
	}
	configFile := ""
	if env == "prod" {
		configFile = "config_prod.yaml"
	} else {
		configFile = "config_dev.yaml"
	}

	file, err := os.ReadFile(configFile)
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	err = yaml.Unmarshal(file, &AppConfig)
	if err != nil {
		log.Fatalf("Error unmarshalling config: %v", err)
	}
}
