package middlewares

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"veripTest/utils"
)

func AuthConfirmMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := ctx.GetHeader("Authorization")
		jwt, b := utils.ValidJWT(token)
		if b && jwt != -1 {
			ctx.Set("id", jwt)
		} else {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"data":    nil,
				"message": "验证jwt失败",
			})
			ctx.Abort()
			return
		}
		ctx.Next()
	}
}
