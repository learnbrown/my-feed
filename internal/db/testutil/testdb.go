package dbtest

import (
	"os"
	"strings"
	"testing"

	"my_feed/internal/config"
	"my_feed/internal/db"

	"gorm.io/gorm"
)

func SetupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	if os.Getenv("RUN_MYSQL_TESTS") != "1" {
		t.Skip("set RUN_MYSQL_TESTS=1 to run MySQL integration tests")
	}

	cfg := config.DatabaseConfig{
		Host:     getenv("MYSQL_HOST", "localhost"),
		Port:     3306,
		User:     getenv("MYSQL_USER", "dev_user"),
		Password: getenv("MYSQL_PASSWORD", "qwerdf"),
		DBName:   getenv("MYSQL_DBNAME", "myfeed_test"),
	}

	if !strings.HasSuffix(cfg.DBName, "_test") {
		t.Fatalf("refuse to run tests on non-test database: %s", cfg.DBName)
	}

	sqlDB, err := db.NewDB(cfg)
	if err != nil {
		t.Fatalf("failed to connect test db: %v", err)
	}

	if err := db.AutoMigrate(sqlDB); err != nil {
		t.Fatalf("failed to auto migrate test db: %v", err)
	}

	t.Cleanup(func() {
		sqlDB2, err := sqlDB.DB()
		if err == nil {
			_ = sqlDB2.Close()
		}
	})

	return sqlDB
}

func CleanTestDB(t *testing.T, sqlDB *gorm.DB) {
	t.Helper()

	tables := []string{
		"messages",
		"comments",
		"likes",
		"follows",
		"video_tags",
		"tags",
		"videos",
		"accounts",
	}

	if err := sqlDB.Exec("SET FOREIGN_KEY_CHECKS = 0").Error; err != nil {
		t.Fatalf("failed to disable foreign key checks: %v", err)
	}

	for _, table := range tables {
		if err := sqlDB.Exec("TRUNCATE TABLE " + table).Error; err != nil {
			t.Fatalf("failed to truncate %s: %v", table, err)
		}
	}

	if err := sqlDB.Exec("SET FOREIGN_KEY_CHECKS = 1").Error; err != nil {
		t.Fatalf("failed to enable foreign key checks: %v", err)
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
