package db

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	mysqlerr "github.com/go-sql-driver/mysql"
	"github.com/mattn/go-sqlite3"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

/*
使用环境变量控制使用的数据库
默认使用sqlite
DB_DRIVER=mysql \
MYSQL_USER=dev_user \
MYSQL_PASSWORD=qwerdf \
MYSQL_HOST=127.0.0.1 \
MYSQL_PORT=3306 \
MYSQL_DATABASE=db001 \
go run ./cmd
*/

var ErrRecordNotFound = gorm.ErrRecordNotFound

func InitDB() (db *gorm.DB) {
	driver := strings.ToLower(strings.TrimSpace(getEnv("DB_DRIVER", "sqlite")))

	var (
		err error
		dsn string
	)

	switch driver {
	case "sqlite", "sqlite3":
		log.Println("database settings: sqlite3")
		dsn = firstNonEmpty(os.Getenv("DB_DSN"), os.Getenv("SQLITE_DSN"), ".run/database/data.db")
		db, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	case "mysql":
		log.Println("database settings: mysql")
		dsn = firstNonEmpty(os.Getenv("DB_DSN"), buildMySQLDSN())
		db, err = gorm.Open(gormmysql.Open(dsn), &gorm.Config{})
	default:
		panic(fmt.Sprintf("unsupported DB_DRIVER: %s", driver))
	}

	if err != nil {
		panic(err)
	}

	return db
}

func buildMySQLDSN() string {
	user := getEnv("MYSQL_USER", "dev_user")
	password := getEnv("MYSQL_PASSWORD", "qwerdf")
	host := getEnv("MYSQL_HOST", "127.0.0.1")
	port := getEnv("MYSQL_PORT", "3306")
	database := getEnv("MYSQL_DATABASE", "myfeed")
	params := getEnv("MYSQL_PARAMS", "charset=utf8mb4&parseTime=True&loc=Local")

	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?%s", user, password, host, port, database, params)
}

func getEnv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

// IsDuplicateKeyError reports whether err is a unique or primary key conflict.
func IsDuplicateKeyError(err error) bool {
	var mysqlErr *mysqlerr.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		return true
	}

	var sqliteErr sqlite3.Error
	if errors.As(err, &sqliteErr) {
		return sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique ||
			sqliteErr.ExtendedCode == sqlite3.ErrConstraintPrimaryKey
	}
	return false
}
