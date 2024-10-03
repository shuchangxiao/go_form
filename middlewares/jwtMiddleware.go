package middlewares

import (
	"github.com/gin-gonic/gin"
	"veripTest/constant"
	"veripTest/global"
	"veripTest/utils"
)

func AuthConfirmMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := ctx.GetHeader("Authorization")
		jwt, b := utils.ValidJWT(token)
		if b && jwt != -1 {
			ctx.Set(constant.UseId, jwt)
		} else {
			global.BusinessErr(ctx, "验证jwt失败")
			return
		}
		ctx.Next()
	}
}
