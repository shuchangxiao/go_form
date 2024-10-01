package controller

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/streadway/amqp"
	"golang.org/x/exp/rand"
	"log"
	"net/http"
	"strconv"
	"time"
	"veripTest/config"
	"veripTest/constant"
	"veripTest/global"
	"veripTest/model"
	"veripTest/utils"
)

// todo 未判断是否已经发送过邮件
func GetCode(ctx *gin.Context) {
	var input struct {
		Email  string `json:"email"`
		Status int    `json:"status"`
	}
	if !global.InitPredicate(ctx, &input) {
		return
	}
	if !global.IsValidEmail(input.Email) {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"data":    nil,
			"message": "邮箱格式错误",
		})
		ctx.Abort()
		return
	}
	if input.Status != 1 && input.Status != 2 && input.Status != 3 {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"data":    nil,
			"message": "未知状态",
		})
		ctx.Abort()
		return
	}
	exists, err1 := global.Redis.Exists(constant.VerifyCode + input.Email).Result()
	if err1 != nil {
		// 处理错误
		global.FailOnErr(ctx)
	}
	if exists > 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"data":    nil,
			"message": "验证码已经发送，请勿重复点击",
		})
		return
	}
	rand.Seed(uint64(time.Now().UnixNano()))
	randomInt := rand.Intn(899999)
	randomInt += 100000
	//redis存储随机参数验证码
	err := global.Redis.Set(constant.VerifyCode+input.Email, randomInt, 5*time.Minute).Err()
	if err != nil {
		global.FailOnErr(ctx)
		log.Printf("redis中存储键值：%v", err)
		return
	}
	//todo:: rabbitmq发送邮件
	sendEmail := model.SendEmail{
		Email:  input.Email,
		Status: input.Status,
		Code:   randomInt,
	}
	sendEmailBytes, err := json.Marshal(sendEmail)
	if err != nil {
		global.FailOnErr(ctx)
		log.Printf("输入格式化错误：%v", err)
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
	ctx.JSON(http.StatusOK, gin.H{
		"data":    nil,
		"message": "邮件已经发送",
	})
}
func ForgetPassword(ctx *gin.Context) {
	var input struct {
		Password string `json:"password"`
		Email    string `json:"email"`
		Code     int    `json:"code"`
	}
	if !global.InitPredicate(ctx, &input) {
		return
	}
	if !global.IsValidEmail(input.Email) {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"data":    nil,
			"message": "邮箱格式错误",
		})
		return
	}
	code, err := global.Redis.Get(constant.VerifyCode + input.Email).Result()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"data":    nil,
			"message": "请先获取验证码",
		})
		ctx.Abort()
		return

	}
	icode, err := strconv.Atoi(code)
	fmt.Println(icode)
	if err != nil {
		global.FailOnErr(ctx)
		log.Printf("从redis中获取失败：%v", err)
		return
	}
	if !(icode == input.Code) {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"data":    nil,
			"message": "验证码错误",
		})
		ctx.Abort()
		return
	}
	var findUser model.User
	err = global.Db.Where("email=?", input.Email).First(&findUser).Error
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"data":    nil,
			"message": "邮箱还未被注册",
		})
		ctx.Abort()
		return
	}
	passwd, err := utils.EncodePasswd(input.Password)
	if err != nil {
		global.FailOnErr(ctx)
		log.Printf("使用hash算法加密错误：%v", err)
		return
	}
	findUser.Password = passwd
	err = global.Db.Updates(&findUser).Error
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"data":    nil,
			"message": "内部错误，请联系管理员",
		})
		ctx.Abort()
		log.Printf("更新用户信息失败：%v", err)
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
		Username string `json:"username"`
		Password string `json:"password"`
		Email    string `json:"email"`
		Code     int    `json:"code"`
	}
	if !global.InitPredicate(ctx, &input) {
		return
	}
	if !global.IsValidEmail(input.Email) {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"data":    nil,
			"message": "邮箱格式错误",
		})
		return
	}
	code, err := global.Redis.Get(constant.VerifyCode + input.Email).Result()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"data":    nil,
			"message": "请先获取验证码",
		})
		ctx.Abort()
		return

	}
	icode, err := strconv.Atoi(code)
	fmt.Println(icode)
	if err != nil {
		global.FailOnErr(ctx)
		log.Printf("从redis中获取失败：%v", err)
		return
	}
	if !(icode == input.Code) {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"data":    nil,
			"message": "验证码错误",
		})
		ctx.Abort()
		return
	}
	var findUser model.User
	err = global.Db.Where("username=?", input.Username).First(&findUser).Error
	if err == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"data":    nil,
			"message": "用户名已经存在",
		})
		ctx.Abort()
		return
	}
	err = global.Db.Where("email=?", input.Email).First(&findUser).Error
	if err == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"data":    nil,
			"message": "邮箱已经被注册",
		})
		ctx.Abort()
		return
	}
	passwd, err := utils.EncodePasswd(input.Password)
	if err != nil {
		global.FailOnErr(ctx)
		log.Printf("使用hash算法加密错误：%v", err)
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
		ctx.Abort()
		log.Printf("创建新用户失败：%v", err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"data":    nil,
		"message": "创建新用户成功",
	})
	global.Redis.Del(constant.VerifyCode + input.Email)
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
		ctx.Abort()
		return
	}
	if !utils.EqualPasswd(user.Password, input.Password) {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"data":    nil,
			"message": "用户名或密码错误",
		})
		ctx.Abort()
		return
	}
	jwt, err := utils.CreateJWT(user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"data":    nil,
			"message": "内部错误，请联系管理员",
		})
		ctx.Abort()
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"token":   jwt,
		"exp":     time.Now().Add(time.Duration(config.Cf.Jwt.Exptime) * time.Hour),
		"message": nil,
	})
}
