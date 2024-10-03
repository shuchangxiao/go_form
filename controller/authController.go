package controller

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/streadway/amqp"
	"golang.org/x/exp/rand"
	"net/http"
	"strconv"
	"time"
	"veripTest/config"
	"veripTest/constant"
	"veripTest/global"
	"veripTest/model"
	"veripTest/utils"
)

func GetCode(ctx *gin.Context) {
	var input struct {
		Email  string `json:"email" validate:"required,email"`
		Status int    `json:"status"  validate:"required,gt=1,lt=5"`
	}
	if !global.InitPredicate(ctx, &input) {
		return
	}
	if input.Status != 1 && input.Status != 2 && input.Status != 3 {
		global.BusinessErr(ctx, "未知状态")
		return
	}
	exists, err1 := global.Redis.Exists(constant.VerifyCode + input.Email).Result()
	if err1 != nil {
		// 处理错误
		global.FailOnErr(ctx, constant.RedisGetKeyErr, err1)
		return
	}
	if exists > 0 {
		global.BusinessErr(ctx, "验证码已经发送，请勿重复点击")
		return
	}
	rand.Seed(uint64(time.Now().UnixNano()))
	randomInt := rand.Intn(899999)
	randomInt += 100000
	//redis存储随机参数验证码
	err := global.Redis.Set(constant.VerifyCode+input.Email, randomInt, 5*time.Minute).Err()
	if err != nil {
		global.FailOnErr(ctx, constant.RedisSetKeyErr, err)
		return
	}
	sendEmail := model.SendEmail{
		Email:  input.Email,
		Status: input.Status,
		Code:   randomInt,
	}
	sendEmailBytes, err := json.Marshal(sendEmail)
	if err != nil {
		global.FailOnErr(ctx, constant.JsonMarshalErr, err)
		return
	}
	err = global.Channel.Publish(
		"",                         // exchange
		global.SendEmailRoutineKey, // routing key
		false,                      // mandatory
		false,                      // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        sendEmailBytes,
		})
	if err != nil {
		global.FailOnErr(ctx, constant.RabbitmqSendErr, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"data":    nil,
		"message": "邮件已经发送",
	})
}
func ForgetPassword(ctx *gin.Context) {
	var input struct {
		Password string `json:"password"  validate:"required,min=6,max=30"`
		Email    string `json:"email" validate:"required,email"`
		Code     int    `json:"code" validate:"required,gt=100000,lt=999999"`
	}
	if !global.InitPredicate(ctx, &input) {
		return
	}
	code, err := global.Redis.Get(constant.VerifyCode + input.Email).Result()
	if err != nil {
		global.BusinessErr(ctx, "请先获取验证码")
		return

	}
	icode, err := strconv.Atoi(code)
	fmt.Println(icode)
	if err != nil {
		global.FailOnErr(ctx, constant.RedisGetKeyErr, err)
		return
	}
	if icode != input.Code {
		global.BusinessErr(ctx, "验证码错误")
		return
	}
	var findUser model.User
	err = global.Db.Where("email=?", input.Email).First(&findUser).Error
	if err != nil {
		global.BusinessErr(ctx, "邮箱未被注册")
		return
	}
	passwd, err := utils.EncodePasswd(input.Password)
	if err != nil {
		global.FailOnErr(ctx, constant.HashEncodeErr, err)
		return
	}
	findUser.Password = passwd
	err = global.Db.Updates(&findUser).Error
	if err != nil {
		global.FailOnErr(ctx, constant.MysqlUpdateErr, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"data":    nil,
		"message": "跟新密码成功",
	})
	global.Redis.Del(constant.VerifyCode + input.Email)
}
func Register(ctx *gin.Context) {
	var input struct {
		Username string `json:"username"  validate:"required,min=1,max=15"`
		Password string `json:"password"  validate:"required,min=6,max=30"`
		Email    string `json:"email" validate:"required,email"`
		Code     int    `json:"code" validate:"required,gt=100000,lt=999999"`
	}
	if !global.InitPredicate(ctx, &input) {
		return
	}
	code, err := global.Redis.Get(constant.VerifyCode + input.Email).Result()
	if err != nil {
		global.BusinessErr(ctx, "请先获取验证码")
		return
	}
	icode, err := strconv.Atoi(code)
	if err != nil {
		global.FailOnErr(ctx, constant.RedisGetKeyErr, err)
		return
	}
	if !(icode == input.Code) {
		global.BusinessErr(ctx, "验证码错误")
		return
	}
	var findUser model.User
	err = global.Db.Where("username=?", input.Username).First(&findUser).Error
	if err == nil {
		global.BusinessErr(ctx, "用户名已经存在")
		return
	}
	err = global.Db.Where("email=?", input.Email).First(&findUser).Error
	if err == nil {
		global.BusinessErr(ctx, "邮箱已经被注册")
		return
	}
	passwd, err := utils.EncodePasswd(input.Password)
	if err != nil {
		global.FailOnErr(ctx, constant.HashEncodeErr, err)
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
		global.FailOnErr(ctx, constant.MysqlUCreateErr, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"data":    nil,
		"message": "创建新用户成功",
	})
	global.Redis.Del(constant.VerifyCode + input.Email)
}

// todo ::未增加邮箱登录

func Login(ctx *gin.Context) {
	var input struct {
		Username string `json:"username"  validate:"required,min=1,max=15"`
		Password string `json:"password"  validate:"required,min=6,max=30"`
	}
	if !global.InitPredicate(ctx, &input) {
		return
	}
	var user model.User
	err := global.Db.Where("username=?", input.Username).First(&user).Error
	if err != nil {
		global.BusinessErr(ctx, "用户名不存在")
		return
	}
	if !utils.EqualPasswd(user.Password, input.Password) {
		global.BusinessErr(ctx, "密码错误")
		return
	}
	jwt, err := utils.CreateJWT(user.ID)
	if err != nil {
		global.FailOnErr(ctx, constant.JWTCreateErr, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"token":   jwt,
		"exp":     time.Now().Add(time.Duration(config.Cf.Jwt.Exptime) * time.Hour),
		"message": nil,
	})
}
