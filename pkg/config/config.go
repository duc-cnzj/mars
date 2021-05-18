package config

import (
	"log"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	AppPort        string
	Debug          bool
	LogChannel     string
	ProfileEnabled bool

	// mysql
	DBHost     string
	DBPort     string
	DBUsername string
	DBPassword string
	DBDatabase string
}

func Init(cfgFile string) *Config {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		dir, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}

		viper.AddConfigPath(dir)
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(err)
	}

	cfg := &Config{
		AppPort:        viper.GetString("app_port"),
		Debug:          viper.GetBool("debug"),
		LogChannel:     viper.GetString("log_channel"),
		ProfileEnabled: viper.GetBool("profile_enabled"),
		DBHost:         viper.GetString("db_host"),
		DBPort:         viper.GetString("db_port"),
		DBUsername:     viper.GetString("db_username"),
		DBPassword:     viper.GetString("db_password"),
		DBDatabase:     viper.GetString("db_database"),
	}

	return cfg
}
