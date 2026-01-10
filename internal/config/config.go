package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	MinIO    MinIOConfig    `mapstructure:"minio"`
	LLM      LLMConfig      `mapstructure:"llm"`
	SMTP     SMTPConfig     `mapstructure:"smtp"`
	Asynq    AsynqConfig    `mapstructure:"asynq"`
	Nacos    NacosConfig    `mapstructure:"nacos"`
}

type ServerConfig struct {
	Port         int           `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	Mode         string        `mapstructure:"mode"`
}

type DatabaseConfig struct {
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	User            string `mapstructure:"user"`
	Password        string `mapstructure:"password"`
	DBName          string `mapstructure:"dbname"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type MinIOConfig struct {
	Endpoint        string `mapstructure:"endpoint"`
	AccessKey       string `mapstructure:"access_key"`
	SecretKey       string `mapstructure:"secret_key"`
	Bucket          string `mapstructure:"bucket"`
	UseSSL          bool   `mapstructure:"use_ssl"`
	Region          string `mapstructure:"region"`
	PresignedExpiry int    `mapstructure:"presigned_expiry"`
}

type LLMConfig struct {
	BaseURL    string        `mapstructure:"base_url"`
	APIKey     string        `mapstructure:"api_key"`
	Model      string        `mapstructure:"model"`
	Timeout    time.Duration `mapstructure:"timeout"`
	MaxRetries int           `mapstructure:"max_retries"`
}

type SMTPConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	From     string `mapstructure:"from"`
}

type AsynqConfig struct {
	RedisAddr     string         `mapstructure:"redis_addr"`
	RedisPassword string         `mapstructure:"redis_password"`
	RedisDB       int            `mapstructure:"redis_db"`
	Concurrency   int            `mapstructure:"concurrency"`
	Queues        map[string]int `mapstructure:"queues"`
}

type NacosConfig struct {
	Addr      string `mapstructure:"addr"`
	Port      uint64 `mapstructure:"port"`
	Namespace string `mapstructure:"namespace"`
	Group     string `mapstructure:"group"`
	DataID    string `mapstructure:"data_id"`
	Username  string `mapstructure:"username"`
	Password  string `mapstructure:"password"`
	LogDir    string `mapstructure:"log_dir"`
	LogLevel  string `mapstructure:"log_level"`
}

func Load(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if cfg.Nacos.Addr != "" {
		if err := loadFromNacos(&cfg); err != nil {
			log.Printf("Warning: failed to load config from Nacos: %v, using local config", err)
		}
	}

	setDefaults(&cfg)

	return &cfg, nil
}

func loadFromNacos(cfg *Config) error {
	sc := []constant.ServerConfig{
		*constant.NewServerConfig(cfg.Nacos.Addr, cfg.Nacos.Port),
	}

	cc := constant.ClientConfig{
		NamespaceId:         cfg.Nacos.Namespace,
		TimeoutMs:           5000,
		NotLoadCacheAtStart: true,
		LogDir:              cfg.Nacos.LogDir,
		LogLevel:            cfg.Nacos.LogLevel,
	}

	if cfg.Nacos.Username != "" {
		cc.Username = cfg.Nacos.Username
		cc.Password = cfg.Nacos.Password
	}

	client, err := clients.NewConfigClient(
		vo.WithNacos(sc...),
		vo.WithClientConfig(cc),
	)
	if err != nil {
		return fmt.Errorf("failed to create Nacos client: %w", err)
	}

	content, err := client.GetConfig(vo.ConfigParam{
		DataId: cfg.Nacos.DataID,
		Group:  cfg.Nacos.Group,
	})
	if err != nil {
		return fmt.Errorf("failed to get config from Nacos: %w", err)
	}

	if content == "" {
		return fmt.Errorf("config content is empty")
	}

	var nacosConfig Config
	if err := json.Unmarshal([]byte(content), &nacosConfig); err != nil {
		return fmt.Errorf("failed to unmarshal Nacos config: %w", err)
	}

	mergeConfig(cfg, &nacosConfig)

	log.Printf("Successfully loaded config from Nacos: %s", cfg.Nacos.DataID)

	return nil
}

func mergeConfig(cfg *Config, nacosCfg *Config) {
	if nacosCfg.Database.Host != "" {
		cfg.Database = nacosCfg.Database
	}
	if nacosCfg.Redis.Host != "" {
		cfg.Redis = nacosCfg.Redis
	}
	if nacosCfg.MinIO.Endpoint != "" {
		cfg.MinIO = nacosCfg.MinIO
	}
	if nacosCfg.LLM.BaseURL != "" {
		cfg.LLM = nacosCfg.LLM
	}
	if nacosCfg.SMTP.Host != "" {
		cfg.SMTP = nacosCfg.SMTP
	}
	if nacosCfg.Asynq.RedisAddr != "" {
		cfg.Asynq = nacosCfg.Asynq
	}
}

func setDefaults(cfg *Config) {
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	if cfg.Server.ReadTimeout == 0 {
		cfg.Server.ReadTimeout = 30 * time.Second
	}
	if cfg.Server.WriteTimeout == 0 {
		cfg.Server.WriteTimeout = 30 * time.Second
	}
	if cfg.Server.Mode == "" {
		cfg.Server.Mode = "debug"
	}

	if cfg.Database.MaxOpenConns == 0 {
		cfg.Database.MaxOpenConns = 100
	}
	if cfg.Database.MaxIdleConns == 0 {
		cfg.Database.MaxIdleConns = 10
	}
	if cfg.Database.ConnMaxLifetime == 0 {
		cfg.Database.ConnMaxLifetime = 3600
	}

	if cfg.MinIO.PresignedExpiry == 0 {
		cfg.MinIO.PresignedExpiry = 3600
	}

	if cfg.LLM.Timeout == 0 {
		cfg.LLM.Timeout = 15 * time.Second
	}
	if cfg.LLM.MaxRetries == 0 {
		cfg.LLM.MaxRetries = 2
	}

	if cfg.Asynq.Concurrency == 0 {
		cfg.Asynq.Concurrency = 10
	}
	if cfg.Asynq.Queues == nil {
		cfg.Asynq.Queues = map[string]int{
			"critical": 6,
			"default":  3,
			"low":      1,
		}
	}

	if cfg.Nacos.Group == "" {
		cfg.Nacos.Group = "DEFAULT_GROUP"
	}
	if cfg.Nacos.DataID == "" {
		cfg.Nacos.DataID = "xiaohongshu-image"
	}
	if cfg.Nacos.Namespace == "" {
		cfg.Nacos.Namespace = "public"
	}
	if cfg.Nacos.LogDir == "" {
		cfg.Nacos.LogDir = "/tmp/nacos/log"
	}
	if cfg.Nacos.LogLevel == "" {
		cfg.Nacos.LogLevel = "info"
	}
}

func GetEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
