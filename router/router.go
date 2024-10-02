package router

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"veripTest/controller"
	"veripTest/middlewares"
)

func SetRouter() *gin.Engine {
	engine := gin.Default()

	auth := engine.Group("/api/auth")
	{
		auth.POST("/login", controller.Login)
		auth.POST("/register", controller.Register)
		auth.POST("/getCode", controller.GetCode)
		auth.POST("/forget", controller.ForgetPassword)
	}
	api := engine.Group("/api")
	api.Use(middlewares.AuthConfirmMiddleware())
	{
		form := api.Group("/form")
		{
			form.POST("/weather", controller.GetHeFengWeather)
			form.GET("/test", func(context *gin.Context) {
				context.JSON(http.StatusOK, gin.H{
					"code":    http.StatusOK,
					"message": "测试案例成功",
				})
			})
		}
	}

	return engine
}
