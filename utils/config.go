package utils

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	DBdriver                   string        `mapstructure:"DB_DRIVER"`
	DBsource                   string        `mapstructure:"DB_SOURCE"`
	ServerAddress              string        `mapstructure:"SERVER_ADDRESS"`
	TokenSymmetricKey          string        `mapstructure:"TOKEN_SYMMETRIC_KEY"`
	AccessTokenDuration        time.Duration `mapstructure:"ACCESS_TOKEN_DURATION"`
	RefreshTokenDuration       time.Duration `mapstructure:"REFRESH_TOKEN_DURATION"`
	PasswordResetTokenDuration time.Duration `mapstructure:"PASSWORD_RESET_TOKEN_DURATION"`
	VerifyEmailTokenDuration   time.Duration `mapstructure:"VERIFY_EMAIL_TOKEN_DURATION"`
	RedisAddress               string        `mapstructure:"REDIS_ADDRESS"`
	EmailSenderName            string        `mapstructure:"EMAIL_SENDER_NAME"`
	EmailSenderAddress         string        `mapstructure:"EMAIL_SENDER_ADDRESS"`
	EmailSenderPassword        string        `mapstructure:"EMAIL_SENDER_PASSWORD"`
}

func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")
	viper.AutomaticEnv()
	err = viper.ReadInConfig()
	if err != nil {
		return
	}
	err = viper.Unmarshal(&config)
	return
}
