package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

var configPath string

func init() {
	flag.StringVar(&configPath, "config", "", "agent config path")
}

type Config struct {
	S3   S3    `mapstructure:"s3" validate:"required"`
	Htts Https `mapstructure:"https" validate:"required"`
	LxdSocket string  `mapstructure:"lxdSocket" validate:"required"`
}

type S3 struct {
	Key        string `mapstructure:"key" validate:"required"`
	Secret     string `mapstructure:"secret" validate:"required"`
	Region     string `mapstructure:"region" validate:"required"`
	Endpoint   string `mapstructure:"endpoint" validate:"required"`
	Bucket     string `mapstructure:"bucket" validate:"required"`
	PartSizeMB uint   `mapstructure:"partSizeMB" validate:"required"`
}

type Https struct {
	Port        string `mapstructure:"port" validate:"required"`
	TlsKeyFile  string `mapstructure:"tlsKeyFile"`
	TlsCertFiel string `mapstructure:"tlsCertFile"`
	Username    string `mapstructure:"username"`
	Password    string `mapstructure:"password"`
}

func InitConfig() (*Config, error) {
	if configPath == "" {
		configPathFromEnv := os.Getenv("CONFIG_PATH")
		if configPathFromEnv != "" {
			configPath = configPathFromEnv
		} else {
			getwd, err := os.Getwd()
			if err != nil {
				return nil, errors.Wrap(err, "os.Getwd")
			}
			configPath = fmt.Sprintf("%s/config.yaml", getwd)
		}
	}

	cfg := &Config{}

	viper.SetConfigType("yaml")
	viper.SetConfigFile(configPath)

	if err := viper.ReadInConfig(); err != nil {
		return nil, errors.Wrap(err, "viper.ReadInConfig")
	}

	if err := viper.Unmarshal(cfg); err != nil {
		return nil, errors.Wrap(err, "viper.Unmarshal")
	}

	return cfg, nil
}
