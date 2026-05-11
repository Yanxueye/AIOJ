package config

import (
	"fmt"
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
	Mode string `yaml:"mode"`
}

type JWTConfig struct {
	Secret      string `yaml:"secret"`
	ExpireHours int    `yaml:"expire_hours"`
}

type MySQLConfig struct {
	DSN         string `yaml:"dsn"`
	MaxIdle     int    `yaml:"max_idle"`
	MaxOpen     int    `yaml:"max_open"`
	AutoMigrate bool   `yaml:"auto_migrate"`
	Seed        bool   `yaml:"seed"`
}

type RabbitMQConfig struct {
	URL         string `yaml:"url"`
	SubmitQueue string `yaml:"submit_queue"`
	Enabled     bool   `yaml:"enabled"`
}

type JudgerConfig struct {
	GRPCAddr       string `yaml:"grpc_addr"`
	TimeoutSeconds int    `yaml:"timeout_seconds"`
}

type RateLimitConfig struct {
	SubmitPerMinute int `yaml:"submit_per_minute"`
	SubmitBurst     int `yaml:"submit_burst"`
}

type AIConfig struct {
	Enabled        bool   `yaml:"enabled"`
	Endpoint       string `yaml:"endpoint"`
	APIKey         string `yaml:"api_key"`
	Model          string `yaml:"model"`
	TimeoutSeconds int    `yaml:"timeout_seconds"`
}

type Config struct {
	Server    ServerConfig    `yaml:"server"`
	JWT       JWTConfig       `yaml:"jwt"`
	MySQL     MySQLConfig     `yaml:"mysql"`
	RabbitMQ  RabbitMQConfig  `yaml:"rabbitmq"`
	Judger    JudgerConfig    `yaml:"judger"`
	RateLimit RateLimitConfig `yaml:"ratelimit"`
	AI        AIConfig        `yaml:"ai"`
}

var (
	globalCfg *Config
	once      sync.Once
)

// Load reads the YAML file into the global singleton.
func Load(path string) (*Config, error) {
	var err error
	once.Do(func() {
		raw, readErr := os.ReadFile(path)
		if readErr != nil {
			err = fmt.Errorf("read config: %w", readErr)
			return
		}
		cfg := &Config{}
		if yErr := yaml.Unmarshal(raw, cfg); yErr != nil {
			err = fmt.Errorf("parse config: %w", yErr)
			return
		}
		globalCfg = cfg
	})
	return globalCfg, err
}

// Get returns the loaded configuration. Panics if Load was not called first.
func Get() *Config {
	if globalCfg == nil {
		panic("config not loaded: call config.Load first")
	}
	return globalCfg
}
