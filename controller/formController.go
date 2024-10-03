package controller

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
	"unsafe"
	"veripTest/constant"
	"veripTest/global"
	"veripTest/model"
	"veripTest/utils"
)

func CreateTopic(ctx *gin.Context) {
	var input struct {
		Code    int                    `json:"code" validate:"required,gt=1,lt=5"`
		Title   string                 `json:"title"  validate:"required,min=1,max=30"`
		Content map[string]interface{} `json:"content" validate:"required"`
	}
	if !global.InitPredicate(ctx, &input) {
		return
	}
	uid := utils.GetUserId(ctx)
	if uid == -1 {
		return
	}
	marshal, err := json.Marshal(input.Content)
	if err != nil {
		global.FailOnErr(ctx, constant.JsonMarshalErr, err)
		return
	}
	topic := model.Topic{
		Title:   input.Title,
		Content: *(*string)(unsafe.Pointer(&marshal)),
		Type_:   input.Code,
		Uid:     uid,
		Top:     0,
	}
	if err := global.Db.Create(&topic).Error; err != nil {
		global.BusinessErr(ctx, "发表文章失败，未知错误")
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"data":    nil,
		"message": "发表文章成功",
	})
	ctx.Abort()
	return
}
func UpdateTopic(ctx *gin.Context) {
	var input struct {
		Id      int                    `json:"id" validate:"required,min=1"`
		Code    int                    `json:"code" validate:"required,gt=1,lt=5"`
		Title   string                 `json:"title"  validate:"required,min=1,max=30"`
		Content map[string]interface{} `json:"content" validate:"required"`
	}
	if !global.InitPredicate(ctx, &input) {
		return
	}
	var topic model.Topic
	err := global.Db.Where("id", input.Id).First(&topic).Error
	if err != nil {
		global.FailOnErr(ctx, constant.MysqlUFindErr, err)
		return
	}
	uid := utils.GetUserId(ctx)
	if uid == -1 {
		return
	}
	if uid != topic.Uid {
		global.BusinessErr(ctx, "您无法修改不是由自己发表的文章")
		return
	}
	topic.Title = input.Title
	marshal, err := json.Marshal(input.Content)
	if err != nil {
		global.FailOnErr(ctx, constant.JsonMarshalErr, err)
		return
	}
	topic.Content = *(*string)(unsafe.Pointer(&marshal))
	global.Db.Updates(topic)
	ctx.JSON(http.StatusOK, gin.H{
		"data":    nil,
		"message": "更新文章成功",
	})
	ctx.Abort()
	return
}
