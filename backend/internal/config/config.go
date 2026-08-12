package config

import (
	"log/slog"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	App     AppConfig     `mapstructure:"app"`
	MySQL   MySQLConfig   `mapstructure:"mysql"`
	Redis   RedisConfig   `mapstructure:"redis"`
	JWT     JWTConfig     `mapstructure:"jwt"`
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
	viper.BindEnv("app.mode", "APP_MODE")

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

	return &config
}
