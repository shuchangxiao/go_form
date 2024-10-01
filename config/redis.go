package config

import (
	"github.com/go-redis/redis"
	"veripTest/global"
)

func InitRedisDB() {
	addr := Cf.Redis.Host + ":" + Cf.Redis.Port
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: Cf.Redis.Password, // no password set
		DB:       Cf.Redis.Db,       // use default DB
	})
	global.Redis = rdb
}
