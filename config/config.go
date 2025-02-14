package config

import (
	"log"
	"os"

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
}

var AppConfig Config

func InitConfig() {
	configFile, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	err = yaml.Unmarshal(configFile, &AppConfig)
	if err != nil {
		log.Fatalf("Error unmarshalling config: %v", err)
	}
}
