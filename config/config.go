package config

import (
	"github.com/spf13/viper"
	"log"
)

type Config struct {
	App struct {
		Name string
		Port string
	}
	DataBase struct {
		Host           string
		Port           string
		Username       string
		Password       string
		Dbname         string
		idleConnection int
		maxConnection  int
		LiveTime       int
	}
	Jwt struct {
		Secret  string
		Exptime uint
	}
	Redis struct {
		Host     string
		Port     string
		Password string
		Db       int
	}
	Rabbitmq struct {
		Username string
		Host     string
		Port     string
		Password string
	}
	Email struct {
		Host     string
		Port     int
		Username string
		Password string
	}
	Weather struct {
		Key string
	}
}

var Cf *Config

func Init() {
	viper.AddConfigPath("./config")
	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("读取文件出现错误：%v", err)
	}
	Cf = &Config{}
	err = viper.Unmarshal(Cf)
	if err != nil {
		log.Fatalf("转换成struct出现错误：%v", err)
	}
	InitDataBase()
	InitRedisDB()
	InitRabbitMQ()
}
