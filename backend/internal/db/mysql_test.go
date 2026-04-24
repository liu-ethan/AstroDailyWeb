package db

import (
	"testing"

	"astrodailyweb/backend/internal/config"
)

func TestNewMySQLInvalidAddr(t *testing.T) {
	_, err := NewMySQL(config.DBConfig{
		Host:            "invalid-host-for-test.localdomain",
		Port:            "3306",
		User:            "root",
		Password:        "",
		Name:            "astro_daily_web",
		MaxOpenConns:    1,
		MaxIdleConns:    1,
		ConnMaxLifetime: 0,
	})
	if err == nil {
		t.Fatal("expected mysql connection error")
	}
}
