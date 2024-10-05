package controller

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"io"
	"mime/multipart"
	"net/http"
	"time"
	"veripTest/config"
	"veripTest/constant"
	"veripTest/global"
	"veripTest/model"
	"veripTest/utils"
)

func UploadImage(ctx *gin.Context) {
	file, err := ctx.FormFile("file")
	if err != nil {
		global.FailOnErr(ctx, constant.ImageUpLoadErr, err)
		return
	}
	open, err := file.Open()
	if err != nil {
		global.FailOnErr(ctx, constant.FileOpenErr, err)
		return
	}
	defer func(open multipart.File) {
		err := open.Close()
		if err != nil {
			global.FailOnErr(ctx, constant.FileCloseErr, err)
			return
		}
	}(open)
	newUUID, err := uuid.NewUUID()
	if err != nil {
		global.FailOnErr(ctx, constant.GenerateUUIDErr, err)
	}
	now := time.Now().Format("20060102")
	imageName := constant.ImagesPath + now + "/" + newUUID.String()
	fmt.Println(file.Filename)
	_, err = global.Minio.PutObject(context.Background(), config.Cf.Minio.BucketName, imageName, open, -1, minio.PutObjectOptions{ContentType: "application/octet-stream"})
	if err != nil {
		global.FailOnErr(ctx, constant.MinioUploadErr, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"data":    imageName,
		"message": nil,
	})
	ctx.Abort()
	return
}
func UploadAvatar(ctx *gin.Context) {
	file, err := ctx.FormFile("file")
	if err != nil {
		global.FailOnErr(ctx, constant.ImageUpLoadErr, err)
		return
	}
	open, err := file.Open()
	if err != nil {
		global.FailOnErr(ctx, constant.FileOpenErr, err)
		return
	}
	defer func(open multipart.File) {
		err := open.Close()
		if err != nil {
			global.FailOnErr(ctx, constant.FileCloseErr, err)
			return
		}
	}(open)
	newUUID, err := uuid.NewUUID()
	if err != nil {
		global.FailOnErr(ctx, constant.GenerateUUIDErr, err)
	}
	avatarPath := constant.AvatarPath + newUUID.String()
	_, err = global.Minio.PutObject(context.Background(), config.Cf.Minio.BucketName, avatarPath, open, -1, minio.PutObjectOptions{ContentType: "application/octet-stream"})
	if err != nil {
		global.FailOnErr(ctx, constant.MinioUploadErr, err)
		return
	}
	dbUser := global.Db.Model(model.User{})
	uid := utils.GetUserId(ctx)
	if uid == 0 {
		return
	}
	var avatar string
	err = dbUser.Where("id", uid).Select("avatar").First(&avatar).Error
	if err != nil {
		global.FailOnErr(ctx, constant.MysqlUFindErr, err)
		return
	}
	if avatar != "" {
		err = global.Minio.RemoveObject(context.Background(), config.Cf.Minio.BucketName, avatar, minio.RemoveObjectOptions{})
		if err != nil {
			global.FailOnErr(ctx, constant.MinioDeleteErr, err)
			return
		}
	}
	err = dbUser.Where("id", uid).Update("avatar", avatarPath).Error
	if err != nil {
		global.FailOnErr(ctx, constant.MysqlUpdateErr, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"data":    avatarPath,
		"message": nil,
	})
	ctx.Abort()
	return
}
func GetImage(ctx *gin.Context) {
	param := ctx.Param("imagePath")
	object, err := global.Minio.GetObject(context.Background(), config.Cf.Minio.BucketName, param, minio.GetObjectOptions{})
	if err != nil {
		global.FailOnErr(ctx, constant.MinioGetErr, err)
		return
	}
	defer func(object *minio.Object) {
		err := object.Close()
		if err != nil {
			global.FailOnErr(ctx, constant.FileCloseErr, err)
		}
	}(object)
	ctx.Header("Cache-Control", "max-age=2592000")
	_, err = io.Copy(ctx.Writer, object)
	if err != nil {
		global.FailOnErr(ctx, constant.ObjectCopyToWriterErr, err)
		return
	}
}
