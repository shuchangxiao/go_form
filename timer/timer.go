package timer

import (
	"errors"
	"fmt"
	"github.com/robfig/cron"
	"gorm.io/gorm"
	"log"
	"strconv"
	"strings"
	"veripTest/constant"
	"veripTest/global"
	"veripTest/model"
)

func InitTimer() {
	job := cron.New()
	err := job.AddFunc("@every 3s", redisConvertTOMysqlCollect)
	if err != nil {
		log.Fatalf("Error adding job:%v", err)
		return
	}
	err = job.AddFunc("@every 3s", redisConvertTOMysqlLike)
	if err != nil {
		log.Fatalf("Error adding job:%v", err)
		return
	}
	job.Start()
	//<-ch
	fmt.Println("timer has break")
}
func redisConvertTOMysqlCollect() {
	redisConvertTOMysql(constant.Collect, &model.Collect{})
}
func redisConvertTOMysqlLike() {
	redisConvertTOMysql(constant.Like, &model.Like{})
}
func redisConvertTOMysql(str string, cl model.CollectAndLike) {
	global.Mutex.Lock()
	all, err := global.Redis.HGetAll(str).Result()
	if len(all) == 0 {
		global.Mutex.Unlock()
		return
	}
	global.Redis.Del(str)
	global.Mutex.Unlock()
	if err != nil {
		log.Fatalf("Error redis job:%v", err)
	}
	for k, v := range all {
		split := strings.Split(k, ":")
		uid, err := strconv.Atoi(split[0])
		if err != nil {
			log.Fatalf("Error redis convent to uid:%v", err)
		}
		tid, err := strconv.Atoi(split[1])
		if err != nil {
			log.Fatalf("Error redis convent to uid:%v", err)
		}
		cl.GetType()
		//查询已经点赞
		result := global.Db.Model(cl.GetType()).Unscoped().Where("tid", tid).Where("uid", uid).Where("deleted_at IS NULL").First(&cl)
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			//为1不管，0的话进行删除
			if v == "0" {
				err := global.Db.Model(cl.GetType()).Delete(cl).Error
				if err != nil {
					log.Fatalf("Error mysql delete:%v", err)
					return
				}
			}

		} else {
			//两种可能，没有这条数据，或者数据真的存在没被删除
			//存在数据但被删除
			if cl.GetDeletedAt().Time.IsZero() {
				if v == "1" {
					err := global.Db.Model(cl.GetType()).Unscoped().Where("tid", tid).Where("uid", uid).Update("deleted_at", nil).Error
					if err != nil {
						log.Fatalf("Error mysql update:%v", err)
						return
					}
				}
			} else {
				//不存在数据，进行创建
				if v == "1" {
					cl.SetUid(uint(uid))
					cl.SetTid(uint(tid))
					err := global.Db.Model(cl.GetType()).Create(cl).Error
					if err != nil {
						log.Fatalf("Error mysql create:%v", err)
						return
					}
				}
			}
		}
	}
}
