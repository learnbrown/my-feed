package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestApplyEnvOverrides(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv("SERVER_PORT", "9090")
	t.Setenv("MYSQL_HOST", "mysql.test")
	t.Setenv("MYSQL_PORT", "3307")
	t.Setenv("MYSQL_USER", "tester")
	t.Setenv("MYSQL_PASSWORD", "secret")
	t.Setenv("MYSQL_DBNAME", "myfeed_test")
	t.Setenv("REDIS_HOST", "redis.test")
	t.Setenv("REDIS_PORT", "6380")
	t.Setenv("REDIS_PASSWORD", "redis-secret")
	t.Setenv("REDIS_DB", "2")
	t.Setenv("REDIS_KEY_PREFIX", "testfeed")
	t.Setenv("REDIS_ENABLED", "false")

	cfg := Config{}
	ApplyEnvOverrides(&cfg)

	if cfg.Server.Port != 9090 {
		t.Fatalf("server port = %d", cfg.Server.Port)
	}
	if cfg.Database.Host != "mysql.test" || cfg.Database.Port != 3307 || cfg.Database.User != "tester" || cfg.Database.Password != "secret" || cfg.Database.DBName != "myfeed_test" {
		t.Fatalf("unexpected database config: %#v", cfg.Database)
	}
	if cfg.Redis.Host != "redis.test" || cfg.Redis.Port != 6380 || cfg.Redis.Password != "redis-secret" || cfg.Redis.DB != 2 || cfg.Redis.KeyPrefix != "testfeed" || cfg.Redis.Enabled {
		t.Fatalf("unexpected redis config: %#v", cfg.Redis)
	}
}

func TestApplyEnvOverridesIgnoresInvalidNumbersAndBoolean(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv("SERVER_PORT", "invalid")
	t.Setenv("MYSQL_PORT", "invalid")
	t.Setenv("REDIS_PORT", "invalid")
	t.Setenv("REDIS_DB", "invalid")
	t.Setenv("REDIS_ENABLED", "invalid")

	cfg := Config{
		Server:   ServerConfig{Port: 8080},
		Database: DatabaseConfig{Port: 3306},
		Redis:    RedisConfig{Port: 6379, DB: 1, Enabled: true},
	}
	ApplyEnvOverrides(&cfg)

	if cfg.Server.Port != 8080 || cfg.Database.Port != 3306 || cfg.Redis.Port != 6379 || cfg.Redis.DB != 1 || !cfg.Redis.Enabled {
		t.Fatalf("invalid overrides changed config: %#v", cfg)
	}
}

func TestLoadParsesYAMLAndAppliesEnvironment(t *testing.T) {
	clearConfigEnv(t)
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	data := []byte("server:\n  port: 8080\ndatabase:\n  host: localhost\n  port: 3306\n  user: dev\n  password: pass\n  dbname: myfeed\nredis:\n  host: localhost\n  port: 6379\n  db: 0\n  key_prefix: myfeed\n  enabled: true\n")
	if err := os.WriteFile(configPath, data, 0o600); err != nil {
		t.Fatalf("write test config: %v", err)
	}
	t.Setenv("SERVER_PORT", "8181")

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Server.Port != 8181 || cfg.Database.DBName != "myfeed" || !cfg.Redis.Enabled {
		t.Fatalf("unexpected config: %#v", cfg)
	}
}

func TestLoadLocalDevUsesDefaultOnlyForMissingFile(t *testing.T) {
	clearConfigEnv(t)
	cfg, usedDefault, err := LoadLoaclDev(filepath.Join(t.TempDir(), "missing.yaml"))
	if err != nil {
		t.Fatalf("LoadLoaclDev() error = %v", err)
	}
	if !usedDefault {
		t.Fatal("expected default config to be used")
	}
	if cfg.Server.Port != 8080 || cfg.Database.DBName != "myfeed" {
		t.Fatalf("unexpected default config: %#v", cfg)
	}
}

func clearConfigEnv(t *testing.T) {
	t.Helper()

	keys := []string{
		"SERVER_PORT",
		"MYSQL_HOST",
		"MYSQL_PORT",
		"MYSQL_USER",
		"MYSQL_PASSWORD",
		"MYSQL_DBNAME",
		"REDIS_HOST",
		"REDIS_PORT",
		"REDIS_PASSWORD",
		"REDIS_DB",
		"REDIS_KEY_PREFIX",
		"REDIS_ENABLED",
	}
	for _, key := range keys {
		t.Setenv(key, "")
	}
}
