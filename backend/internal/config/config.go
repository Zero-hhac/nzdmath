package config

import (
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	App     AppConfig     `mapstructure:"app"`
	MySQL   MySQLConfig   `mapstructure:"mysql"`
	Redis   RedisConfig   `mapstructure:"redis"`
	JWT     JWTConfig     `mapstructure:"jwt"`
	SMTP    SMTPConfig    `mapstructure:"smtp"`
	Storage StorageConfig `mapstructure:"storage"`
}

type AppConfig struct {
	Name string `mapstructure:"name"`
	Port string `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

type MySQLConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type JWTConfig struct {
	Secret      string `mapstructure:"secret"`
	ExpireHours int    `mapstructure:"expire_hours"`
}

type SMTPConfig struct {
	Host     string `mapstructure:"host"`     // 邮件服务器地址
	Port     int    `mapstructure:"port"`     // 端口（如 465/587）
	User     string `mapstructure:"user"`     // 发件账号
	Password string `mapstructure:"password"` // 授权码
	From     string `mapstructure:"from"`     // 发件人地址
}

type StorageConfig struct {
	UploadDir string `mapstructure:"upload_dir"`
}

var GlobalConfig *Config

func LoadConfig() *Config {
	if GlobalConfig != nil {
		return GlobalConfig
	}
	viper.SetConfigName("config")
	if env := os.Getenv("APP_ENV"); env != "" {
		viper.SetConfigName("config." + env)
	}
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("./internal/config")

	viper.BindEnv("mysql.password", "MYSQL_PASSWORD")
	viper.BindEnv("jwt.secret", "JWT_SECRET")
	viper.BindEnv("redis.password", "REDIS_PASSWORD")
	viper.BindEnv("app.mode", "APP_MODE")
	viper.BindEnv("smtp.host", "SMTP_HOST")
	viper.BindEnv("smtp.port", "SMTP_PORT")
	viper.BindEnv("smtp.user", "SMTP_USER")
	viper.BindEnv("smtp.password", "SMTP_PASSWORD")
	viper.BindEnv("smtp.from", "SMTP_FROM")

	if err := viper.ReadInConfig(); err != nil {
		slog.Error("读取配置文件失败", "err", err)
		panic(err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		slog.Error("解析配置文件失败", "err", err)
		panic(err)
	}

	GlobalConfig = &config
	slog.Info("配置加载成功",
		"app", config.App.Name,
		"port", config.App.Port,
		"mode", config.App.Mode)

	validateProductionConfig(&config)

	return &config
}

// 占位符 secret，本地开发配置里也禁止在 release 模式使用
const placeholderSecret = "随便填一个长字符串"

// validateProductionConfig 在 release 模式下强制校验敏感配置，
// 防止带着弱密钥/空密码上线。校验失败直接 panic 拒绝启动，不静默降级。
func validateProductionConfig(c *Config) {
	if c.App.Mode != "release" {
		return
	}
	secret := strings.TrimSpace(c.JWT.Secret)
	if len(secret) < 32 || secret == placeholderSecret {
		slog.Error("生产配置校验失败",
			"reason", "JWT_SECRET 必须为至少 32 位的强随机字符串，请通过环境变量注入")
		panic("production config validation failed: invalid jwt secret")
	}
	if strings.TrimSpace(c.MySQL.Password) == "" {
		slog.Error("生产配置校验失败",
			"reason", "MYSQL_PASSWORD 不能为空，请通过环境变量注入")
		panic("production config validation failed: empty mysql password")
	}
}
