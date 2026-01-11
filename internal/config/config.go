package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server" json:"server"`
	Database DatabaseConfig `mapstructure:"database" json:"database"`
	Redis    RedisConfig    `mapstructure:"redis" json:"redis"`
	MinIO    MinIOConfig    `mapstructure:"minio" json:"minio"`
	LLM      LLMConfig      `mapstructure:"llm" json:"llm"`
	SMTP     SMTPConfig     `mapstructure:"smtp" json:"smtp"`
	Asynq    AsynqConfig    `mapstructure:"asynq" json:"asynq"`
	Nacos    NacosConfig    `mapstructure:"nacos" json:"nacos"`
}

type ServerConfig struct {
	Port         int           `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	Mode         string        `mapstructure:"mode"`
}

type DatabaseConfig struct {
	Host            string `mapstructure:"host" json:"host"`
	Port            int    `mapstructure:"port" json:"port"`
	User            string `mapstructure:"user" json:"user"`
	Password        string `mapstructure:"password" json:"password"`
	DBName          string `mapstructure:"dbname" json:"dbname"`
	MaxOpenConns    int    `mapstructure:"max_open_conns" json:"max_open_conns"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns" json:"max_idle_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime" json:"conn_max_lifetime"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host" json:"host"`
	Port     int    `mapstructure:"port" json:"port"`
	Password string `mapstructure:"password" json:"password"`
	DB       int    `mapstructure:"db" json:"db"`
}

type MinIOConfig struct {
	Endpoint        string `mapstructure:"endpoint" json:"endpoint"`
	AccessKey       string `mapstructure:"access_key" json:"access_key"`
	SecretKey       string `mapstructure:"secret_key" json:"secret_key"`
	Bucket          string `mapstructure:"bucket" json:"bucket"`
	UseSSL          bool   `mapstructure:"use_ssl" json:"use_ssl"`
	Region          string `mapstructure:"region" json:"region"`
	PresignedExpiry int    `mapstructure:"presigned_expiry" json:"presigned_expiry"`
}

type LLMConfig struct {
	BaseURL    string        `mapstructure:"base_url" json:"base_url"`
	APIKey     string        `mapstructure:"api_key" json:"api_key"`
	Model      string        `mapstructure:"model" json:"model"`
	Timeout    time.Duration `mapstructure:"timeout" json:"timeout"`
	MaxRetries int           `mapstructure:"max_retries" json:"max_retries"`
}

type SMTPConfig struct {
	Host     string `mapstructure:"host" json:"host"`
	Port     int    `mapstructure:"port" json:"port"`
	User     string `mapstructure:"user" json:"user"`
	Password string `mapstructure:"password" json:"password"`
	From     string `mapstructure:"from" json:"from"`
}

type AsynqConfig struct {
	RedisAddr     string         `mapstructure:"redis_addr" json:"redis_addr"`
	RedisPassword string         `mapstructure:"redis_password" json:"redis_password"`
	RedisDB       int            `mapstructure:"redis_db" json:"redis_db"`
	Concurrency   int            `mapstructure:"concurrency" json:"concurrency"`
	Queues        map[string]int `mapstructure:"queues" json:"queues"`
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

// findConfigFile 尝试从多个位置查找配置文件
func findConfigFile(configPath string) (string, error) {
	// 首先尝试直接使用提供的路径
	if _, err := os.Stat(configPath); err == nil {
		absPath, _ := filepath.Abs(configPath)
		return absPath, nil
	}

	// 如果失败，尝试从项目根目录查找
	// 通过查找 go.mod 文件来确定项目根目录
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	// 向上查找 go.mod 文件
	dir := wd
	for {
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			// 找到项目根目录，尝试在此查找配置文件
			configFullPath := filepath.Join(dir, configPath)
			if _, err := os.Stat(configFullPath); err == nil {
				absPath, _ := filepath.Abs(configFullPath)
				return absPath, nil
			}
			break
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// 已到达文件系统根目录
			break
		}
		dir = parent
	}

	// 如果都找不到，返回原始路径（让 viper 报错）
	return configPath, nil
}

func Load(configPath string) (*Config, error) {
	// 尝试查找配置文件
	resolvedPath, err := findConfigFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve config file path: %w", err)
	}

	viper.SetConfigFile(resolvedPath)
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

	client, err := clients.NewConfigClient(vo.NacosClientParam{
		ClientConfig:  &cc,
		ServerConfigs: sc,
	})
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
