package controller

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
	"veripTest/config"
	"veripTest/global"
	"veripTest/model"
	"veripTest/utils"
)

func getCode() {

}

// todo::还未进行redis相关配置和验证
func Register(ctx *gin.Context) {
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Email    string `json:"email"`
		Code     int    `json:"code"`
	}
	if !global.InitPredicate(ctx, &input) {
		return
	}
	var findUser model.User
	err := global.Db.Where("username=?", input.Username).First(&findUser).Error
	if err == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"data":    nil,
			"message": "用户名已经存在",
		})
		return
	}
	err = global.Db.Where("email=?", input.Email).First(&findUser).Error
	if err == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"data":    nil,
			"message": "邮箱已经被注册",
		})
		return
	}
	passwd, err := utils.EncodePasswd(input.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"data":    nil,
			"message": "内部错误",
		})
		return
	}
	user := model.User{
		Username: input.Username,
		Password: passwd,
		Email:    input.Email,
		Role:     "user",
	}
	err = global.Db.Create(&user).Error
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"data":    nil,
			"message": "内部错误，请联系管理员",
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"data":    nil,
		"message": "创建新用户成功",
	})
}
func Login(ctx *gin.Context) {
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if !global.InitPredicate(ctx, &input) {
		return
	}
	var user model.User
	err := global.Db.Where("username=?", input.Username).First(&user).Error
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"data":    nil,
			"message": "用户名不存在",
		})
		return
	}
	if !utils.EqualPasswd(user.Password, input.Password) {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"data":    nil,
			"message": "用户名或密码错误",
		})
		return
	}
	jwt, err := utils.CreateJWT(user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"data":    nil,
			"message": "内部错误，请联系管理员",
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"token":   jwt,
		"exp":     time.Now().Add(time.Duration(config.Cf.Jwt.Exptime) * time.Hour),
		"message": nil,
	})
}
