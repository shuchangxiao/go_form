package config

import (
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
	"log"
	"veripTest/timer"
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
	Minio struct {
		Ip              string
		Port            string
		AccessKeyID     string
		SecretAccessKey string
		UseSSL          bool
		BucketName      string
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
	InitMinio()
	timer.InitTimer()
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		err := v.RegisterValidation("typeCheck", typeCheck)
		if err != nil {
			log.Fatalf("%v", err)
			return
		}
	}
}
func typeCheck(fl validator.FieldLevel) bool {
	if fl.Field().String() == "collect" || fl.Field().String() == "like" {
		return true
	}
	return false
}
