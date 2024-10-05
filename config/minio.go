package config

import (
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"log"
	"veripTest/global"
)

func InitMinio() {
	endpoint := Cf.Minio.Ip + ":" + Cf.Minio.Port
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(Cf.Minio.AccessKeyID, Cf.Minio.SecretAccessKey, ""),
		Secure: Cf.Minio.UseSSL,
	})
	if err != nil {
		log.Fatalf("初始化minio时出现错误：%v", err)
	}
	global.Minio = minioClient
}
