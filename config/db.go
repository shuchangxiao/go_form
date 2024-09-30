package config

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"time"
	"veripTest/global"
)

// InitDataBase 初始化数据库连接。
// 该函数根据配置文件中的数据库信息生成数据库连接字符串，尝试建立数据库连接，
// 并配置数据库连接的最大闲置连接数、最大打开连接数以及连接的最大生命周期。
func InitDataBase() {
	// 生成数据库连接字符串。
	dns := fmt.Sprintf("%v:%v@tcp(%v:%v)/%v?charset=utf8mb4&parseTime=True&loc=Local",
		Cf.DataBase.Username, Cf.DataBase.Password, Cf.DataBase.Host, Cf.DataBase.Port, Cf.DataBase.Dbname)

	// 尝试打开数据库连接。
	db, err := gorm.Open(mysql.Open(dns), &gorm.Config{})
	if err != nil {
		// 如果出现错误，记录并终止程序。
		log.Fatalf("连接数据库时出现错误：%v", err)
	}

	// 获取底层的SQL数据库对象。
	sqlDB, err := db.DB()
	if err != nil {
		// 如果出现错误，记录并终止程序。
		log.Fatalf("获取底层的SQL数据库对象时出现错误：%v", err)
	}

	// 配置数据库连接的最大闲置连接数。
	sqlDB.SetMaxIdleConns(Cf.DataBase.idleConnection)
	// 配置数据库连接的最大打开连接数。
	sqlDB.SetMaxOpenConns(Cf.DataBase.maxConnection)
	// 配置数据库连接的最大生命周期。
	sqlDB.SetConnMaxLifetime(time.Duration(Cf.DataBase.LiveTime) * time.Minute)

	// 将数据库对象赋值给全局变量，以便其他地方使用。
	global.Db = db
}
