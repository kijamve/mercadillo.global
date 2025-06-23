package H

import (
	"log"
	"os"
	"sync"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	db   *gorm.DB
	once sync.Once
)

func GetDB() *gorm.DB {
	once.Do(func() {
		var err error
		if os.Getenv("MYSQL_DEBUG") == "true" {
			db, err = gorm.Open(mysql.Open(os.Getenv("MYSQL_CONN")), &gorm.Config{
				Logger: logger.Default.LogMode(logger.Info),
			})
		} else {
			newLogger := logger.New(
				log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
				logger.Config{
					SlowThreshold:             time.Second,   // Slow SQL threshold
					LogLevel:                  logger.Silent, // Log level
					IgnoreRecordNotFoundError: true,          // Ignore ErrRecordNotFound error for logger
					ParameterizedQueries:      true,          // Don't include params in the SQL log
					Colorful:                  false,         // Disable color
				},
			)

			db, err = gorm.Open(mysql.Open(os.Getenv("MYSQL_CONN")), &gorm.Config{
				Logger: newLogger,
			})
		}
		if err != nil {
			panic("failed to connect database")
		}

		pool, _ := db.DB()
		pool.SetConnMaxLifetime(10 * time.Minute)
		pool.SetMaxIdleConns(10)
		pool.SetMaxOpenConns(25)
	})
	return db
}

func DB() *gorm.DB {
	if os.Getenv("MYSQL_DEBUG") == "true" {
		return GetDB().Debug()
	}
	return GetDB()
}
