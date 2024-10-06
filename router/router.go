package router

import (
	"github.com/gin-gonic/gin"
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
		api.POST("/weather", controller.GetHeFengWeather)
		form := api.Group("/form")
		{
			form.POST("/create-topic", controller.CreateTopic)
			form.POST("/update-topic", controller.UpdateTopic)
			form.POST("/delete-topic", controller.DeleteTopic)

			form.POST("/create-comment", controller.CreateComment)
			form.POST("/delete-comment", controller.DeleteComment)
			form.POST("/list-topic", controller.ListTopic)
			form.POST("/list-comments", controller.ListComments)
		}
		image := api.Group("/image")
		{
			image.POST("/cache", controller.UploadImage)
			image.POST("/avatar", controller.UploadAvatar)
		}
	}
	engine.GET("/images/*imagePath", controller.GetImage)
	return engine
}
