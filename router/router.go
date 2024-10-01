package router

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"veripTest/controller"
)

func SetRouter() *gin.Engine {
	engine := gin.Default()
	engine.GET("/test", func(context *gin.Context) {
		context.JSON(http.StatusOK, gin.H{
			"code":    http.StatusOK,
			"message": "测试案例成功",
		})
	})
	api := engine.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/login", controller.Login)
			auth.POST("/register", controller.Register)
			auth.POST("/getCode", controller.GetCode)
			auth.POST("/forget", controller.ForgetPassword)
		}
	}

	return engine
}
