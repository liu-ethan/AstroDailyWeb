package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	App  AppConfig  `yaml:"app"`
	DB   DBConfig   `yaml:"db"`
	JWT  JWTConfig  `yaml:"jwt"`
	SMTP SMTPConfig `yaml:"smtp"`
	LLM  LLMConfig  `yaml:"llm"`
}

type AppConfig struct {
	Name       string        `yaml:"name"`
	Env        string        `yaml:"env"`
	Host       string        `yaml:"host"`
	Port       string        `yaml:"port"`
	LogLevel   string        `yaml:"log_level"`
	ReadTO     time.Duration `yaml:"read_timeout"`
	WriteTO    time.Duration `yaml:"write_timeout"`
	IdleTO     time.Duration `yaml:"idle_timeout"`
	EnableCron bool          `yaml:"enable_cron"`
}

type DBConfig struct {
	Host            string        `yaml:"host"`
	Port            string        `yaml:"port"`
	User            string        `yaml:"user"`
	Password        string        `yaml:"password"`
	Name            string        `yaml:"name"`
	MaxOpenConns    int           `yaml:"max_open_conns"`
	MaxIdleConns    int           `yaml:"max_idle_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
}

type JWTConfig struct {
	Secret        string `yaml:"secret"`
	Issuer        string `yaml:"issuer"`
	ExpireMinutes int    `yaml:"expire_minutes"`
}

type SMTPConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
	User string `yaml:"user"`
	Pass string `yaml:"pass"`
	From string `yaml:"from"`
}

type LLMConfig struct {
	Provider string        `yaml:"provider"`
	APIKey   string        `yaml:"api_key"`
	BaseURL  string        `yaml:"base_url"`
	Model    string        `yaml:"model"`
	Timeout  time.Duration `yaml:"timeout"`
}

// Load 从指定 YAML 文件加载配置并校验关键字段。
// 参数：path - 配置文件路径。
// 返回：Config - 配置对象；error - 加载或校验失败时返回错误。
func Load(path string) (Config, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config file failed: %w", err)
	}

	cfg := Config{}
	if err = yaml.Unmarshal(bytes, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config yaml failed: %w", err)
	}

	if cfg.JWT.Secret == "" {
		return Config{}, fmt.Errorf("jwt.secret is required")
	}
	return cfg, nil
}
