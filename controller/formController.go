package controller

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"reflect"
	"sort"
	"strings"
	"sync"
	"time"
	"unsafe"
	"veripTest/constant"
	"veripTest/global"
	"veripTest/model"
	"veripTest/utils"
)

func CreateTopic(ctx *gin.Context) {
	var input struct {
		Code    int                    `json:"code" validate:"required,gt=1"`
		Title   string                 `json:"title"  validate:"required,min=1,max=30"`
		Content map[string]interface{} `json:"content" validate:"required"`
	}
	if !global.InitPredicate(ctx, &input) {
		return
	}
	uid := utils.GetUserId(ctx)
	if uid == 0 {
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
		Code    int                    `json:"code" validate:"required,gt=1"`
		Title   string                 `json:"title"  validate:"required,min=1,max=30"`
		Content map[string]interface{} `json:"content" validate:"required"`
	}
	if !global.InitPredicate(ctx, &input) {
		return
	}
	var topic model.Topic
	err := global.Db.Where("id", input.Id).First(&topic).Error
	if err != nil && err.Error() == constant.MySqlNotFound {
		global.BusinessErr(ctx, constant.TopicNotFound)
		return
	}
	uid := utils.GetUserId(ctx)
	if uid == 0 {
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
func DeleteTopic(ctx *gin.Context) {
	var input struct {
		Id int `json:"id" validate:"required,min=1"`
	}
	if !global.InitPredicate(ctx, &input) {
		return
	}
	uid := utils.GetUserId(ctx)
	if uid == 0 {
		return
	}
	if !deleteRecord(ctx, input.Id, uid, model.Topic{}, "您无法删除不是自己的评论") {
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"data":    nil,
		"message": "删除文章成功",
	})
	ctx.Abort()
	return
}

// TopicPreviewVO 结构体定义
type TopicPreviewVO struct {
	ID       uint      `json:"id"`
	Type     int       `json:"type"`
	Title    string    `json:"title"`
	Text     string    `json:"text"`
	Image    []string  `json:"image"`
	Time     time.Time `json:"time"`
	UID      uint      `json:"uid,omitempty"`
	Username string    `json:"username"`
	Avatar   string    `json:"avatar"`
	Like     int64     `json:"like"`
	Collect  int64     `json:"collect"`
}

// todo 出现bug，没有code也行
func ListTopic(ctx *gin.Context) {
	var input struct {
		Page int `json:"page" validate:"required,gt=1"`
		Code int `json:"code" validate:"required,gt=0"`
	}
	if !global.InitPredicate(ctx, &input) {
		return
	}
	var topics []model.Topic
	offset := (input.Page - 1) * constant.PageSize
	dbTopic := global.Db.Model(model.Topic{})
	var err error
	if input.Code == 0 {
		err = dbTopic.Offset(offset).Order("created_at desc").Limit(constant.PageSize).Find(&topics).Error
	} else {
		err = dbTopic.Where("type", input.Code).Offset(offset).Limit(constant.PageSize).Find(&topics).Error
	}
	if err != nil {
		global.FailOnErr(ctx, constant.MysqlUFindErr, err)
		return
	}
	var previewTopics = make([]TopicPreviewVO, 0, constant.PageSize)
	var wg sync.WaitGroup
	var mutex sync.Mutex
	for _, topic := range topics {
		wg.Add(1)
		go func(topic model.Topic) {
			defer wg.Done()
			var innerWg sync.WaitGroup
			var user model.User
			var previewTopic TopicPreviewVO
			innerWg.Add(1)
			go func() {
				defer innerWg.Done()
				dbUser := global.Db.Model(model.User{})
				if err := dbUser.Where("id", topic.Uid).First(&user).Error; err != nil {
					global.FailOnErr(ctx, constant.MysqlUFindErr, err)
					return
				}
			}()
			var like, collect int64
			innerWg.Add(1)
			go func() {
				defer innerWg.Done()
				dbLike := global.Db.Model(model.Like{})
				if err := dbLike.Where("tid", topic.ID).Count(&like).Error; err != nil {
					global.FailOnErr(ctx, constant.MysqlUFindErr, err)
					return
				}
			}()
			innerWg.Add(1)
			go func() {
				defer innerWg.Done()
				dbCollect := global.Db.Model(model.Collect{})
				if err := dbCollect.Where("tid", topic.ID).Count(&collect).Error; err != nil {
					global.FailOnErr(ctx, constant.MysqlUFindErr, err)
					return
				}
			}()
			innerWg.Add(1)
			go func() {
				defer innerWg.Done()
				if !shortContentAndSetImage(ctx, &previewTopic, topic.Content) {
					return
				}
			}()
			innerWg.Wait()
			previewTopic.ID = topic.ID
			previewTopic.Type = topic.Type_
			previewTopic.Title = topic.Title
			previewTopic.Time = topic.CreatedAt
			previewTopic.UID = topic.Uid
			previewTopic.Username = user.Username
			previewTopic.Avatar = user.Avatar
			previewTopic.Like = like
			previewTopic.Collect = collect
			mutex.Lock()
			previewTopics = append(previewTopics, previewTopic)
			mutex.Unlock()
		}(topic)
	}
	wg.Wait()
	sort.Slice(previewTopics, func(i, j int) bool {
		return previewTopics[i].Time.After(previewTopics[j].Time)
	})
	ctx.JSON(http.StatusOK, gin.H{
		"data":    previewTopics,
		"message": nil,
	})
}

func CreateComment(ctx *gin.Context) {
	var input struct {
		Tid     uint                   `json:"tid" validate:"required,gt=1"`
		Content map[string]interface{} `json:"content"  validate:"required"`
		Quote   uint                   `json:"quote"  validate:"required,gt=0"`
	}
	if !global.InitPredicate(ctx, &input) {
		return
	}
	topicDb := global.Db.Model(model.Topic{})
	var realTid int
	if err := topicDb.Select("id").Where("id", input.Tid).First(&realTid).Error; err != nil && err.Error() == constant.MySqlNotFound {
		global.BusinessErr(ctx, "不存在此文章或评论已被删除")
		return
	}
	commentDb := global.Db.Model(model.Comment{})
	var realCommentId int
	if err := commentDb.Select("id").Where("id", input.Quote).First(&realCommentId).Error; err != nil && err.Error() == constant.MySqlNotFound {
		global.BusinessErr(ctx, "不存在此评论或评论已被删除")
		return
	}
	uid := utils.GetUserId(ctx)
	if uid == 0 {
		return
	}
	marshal, err := json.Marshal(input.Content)
	if err != nil {
		global.FailOnErr(ctx, constant.JsonMarshalErr, err)
		return
	}
	comment := model.Comment{
		Uid:     uid,
		Tid:     input.Tid,
		Content: *(*string)(unsafe.Pointer(&marshal)),
		Quote:   input.Quote,
	}
	if err := global.Db.Create(&comment).Error; err != nil {
		global.FailOnErr(ctx, constant.MysqlUCreateErr, err)
	}
	ctx.JSON(http.StatusOK, gin.H{
		"data":    "创建评论成功",
		"message": nil,
	})
}
func DeleteComment(ctx *gin.Context) {
	var input struct {
		Id int `json:"id" validate:"required,min=1"`
	}
	if !global.InitPredicate(ctx, &input) {
		return
	}
	uid := utils.GetUserId(ctx)
	if uid == 0 {
		return
	}
	if !deleteRecord(ctx, input.Id, uid, model.Comment{}, "您无法删除不是自己的评论") {
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"data":    nil,
		"message": "删除评论成功",
	})
	ctx.Abort()
	return
}

type CommentVO struct {
	ID       uint   `json:"id"`
	Content  string `json:"content"`
	Time     time.Time
	Quote    string `json:"quote"`
	UID      uint   `json:"uid"`
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
}

// todo 出现pandic,&content出现错误
func ListComments(ctx *gin.Context) {
	var input struct {
		Id   int `json:"id" validate:"required,gt=1"`
		Page int `json:"page" validate:"required,gt=1"`
	}
	if !global.InitPredicate(ctx, &input) {
		return
	}
	offset := (input.Page - 1) * constant.CommentsSize
	var comments []model.Comment
	dbComment := global.Db.Model(model.Comment{})
	if err := dbComment.Offset(offset).Limit(constant.CommentsSize).Order("created_at desc").Find(&comments).Error; err != nil {
		global.FailOnErr(ctx, constant.MysqlUFindErr, err)
		return
	}
	commentsVo := make([]CommentVO, 0, constant.CommentsSize)
	var wg sync.WaitGroup
	var mux sync.Mutex
	for _, comment := range comments {
		wg.Add(1)
		go func(comment model.Comment) {
			var innerWg sync.WaitGroup
			defer wg.Done()
			var commentVO CommentVO
			var user model.User
			var content string
			innerWg.Add(1)
			go func() {
				defer innerWg.Done()
				dbUser := global.Db.Model(model.User{})
				if err := dbUser.Where("id", comment.Uid).First(&user).Error; err != nil {
					global.FailOnErr(ctx, constant.MysqlUFindErr, err)
					return
				}
			}()
			if comment.Quote != 0 {
				innerWg.Add(1)
				go func() {
					defer innerWg.Done()
					if err := dbComment.Where("id", comment.Quote).Select("content").First(&content).Error; err.Error() == constant.MySqlNotFound {
						content = "此评论已被删除"
					}
				}()
			}
			innerWg.Wait()
			commentVO.ID = comment.ID
			commentVO.Time = comment.CreatedAt
			commentVO.Content = comment.Content
			commentVO.Quote = content
			commentVO.UID = user.ID
			commentVO.Username = user.Username
			commentVO.Avatar = user.Avatar
			mux.Lock()
			commentsVo = append(commentsVo, commentVO)
			mux.Unlock()
		}(comment)
	}
	wg.Wait()
	sort.Slice(commentsVo, func(i, j int) bool {
		return commentsVo[i].ID > commentsVo[j].ID
	})
	ctx.JSON(http.StatusOK, gin.H{
		"data":    commentsVo,
		"message": nil,
	})
}
func shortContentAndSetImage(ctx *gin.Context, previewTopic *TopicPreviewVO, content string) bool {
	var all map[string]interface{}
	if err := json.Unmarshal([]byte(content), &all); err != nil {
		global.FailOnErr(ctx, constant.JsonUnMarshalErr, err)
		return false
	}
	// 获取到ops里面的内容
	ops, ok := all["ops"].([]interface{})
	if !ok {
		global.FailOnErr(ctx, constant.JsonUnMarshalErr, errors.New("转换成[]interface类型出现错误"))
		return false
	}
	var builder strings.Builder
	//遍历ops里面的内容
	for _, s := range ops {
		//将里面的内容转换成{insert:??}
		smap, ok := s.(map[string]interface{})
		if !ok {
			global.FailOnErr(ctx, constant.JsonUnMarshalErr, errors.New("转换成map[string]interface类型出现错误"))

		}
		strOrImage := smap["insert"]
		typeOf := reflect.TypeOf(strOrImage)
		switch typeOf.Kind() {
		case reflect.String:
			if builder.Len() <= 30 {
				builder.WriteString(strOrImage.(string))
			}
		case reflect.Map:
			m, ok := strOrImage.(map[string]interface{})
			if !ok {
				global.FailOnErr(ctx, constant.JsonUnMarshalErr, errors.New("转换成map[string]string类型出现错误"))
				return false
			}
			image, ok := m["image"].(string)
			if !ok {
				global.FailOnErr(ctx, constant.JsonUnMarshalErr, errors.New("转换成string类型出现错误"))
				return false
			}
			if len(previewTopic.Image) <= 3 {
				previewTopic.Image = append(previewTopic.Image, image)
			}
		default:
			continue
		}
	}
	previewTopic.Text = builder.String()
	return true
}
func deleteRecord(ctx *gin.Context, id int, uid uint, model interface{}, message string) bool {
	var realUid uint
	db := global.Db.Model(model)
	err := db.Where("id", id).Select("uid").First(&realUid).Error
	if err != nil && err.Error() == constant.MySqlNotFound {
		global.BusinessErr(ctx, constant.TopicNotFound)
		return false
	}
	if realUid != uid {
		global.BusinessErr(ctx, message)
		return false
	}
	err = global.Db.Delete(&model, id).Error
	if err != nil {
		global.FailOnErr(ctx, constant.MysqlUpdateErr, err)
		return false
	}
	return true
}
