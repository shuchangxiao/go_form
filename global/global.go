package global

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/minio/minio-go/v7"
	"github.com/streadway/amqp"
	"gorm.io/gorm"
	"log"
	"net/http"
	"sync"
)

var (
	Db                  *gorm.DB
	Redis               *redis.Client
	Channel             *amqp.Channel
	SendEmailRoutineKey string
	Minio               *minio.Client
	Mutex               sync.Mutex
)

func InitPredicate(ctx *gin.Context, input interface{}) bool {
	if err := ctx.ShouldBindJSON(input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"data":    nil,
			"message": "请输入正确的参数",
		})
		fmt.Printf("%v", err)
		ctx.Abort()
		return false
	}
	return true
}

func FailOnErr(ctx *gin.Context, message string, err error) {
	ctx.JSON(http.StatusInternalServerError, gin.H{
		"data":    nil,
		"message": "内部错误，请联系管理员",
	})
	ctx.Abort()
	log.Fatalf("%s:%v", message, err)
}

func BusinessErr(ctx *gin.Context, message string) {
	ctx.JSON(http.StatusInternalServerError, gin.H{
		"data":    nil,
		"message": message,
	})
	ctx.Abort()
}
