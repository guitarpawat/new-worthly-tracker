package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type App struct {
	Name string `mapstructure:"name"`
	Env  string `mapstructure:"env"`
	DB   DB     `mapstructure:"db"`
	Log  Log    `mapstructure:"log"`
}

type DB struct {
	Path string `mapstructure:"path"`
}

type Log struct {
	Path  string `mapstructure:"path"`
	Level string `mapstructure:"level"`
}

func Load(configPath string) (App, error) {
	cfg := App{}

	v := viper.New()
	v.SetEnvPrefix("WORTHLY")
	v.AutomaticEnv()
	v.SetDefault("name", "Worthly Tracker")
	v.SetDefault("env", "development")
	v.SetDefault("log.level", "INFO")

	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("app")
		v.SetConfigType("yaml")
		v.AddConfigPath("./config")
	}

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return App{}, fmt.Errorf("read config: %w", err)
		}
	}

	if err := v.Unmarshal(&cfg); err != nil {
		return App{}, fmt.Errorf("unmarshal config: %w", err)
	}

	return cfg, nil
}
