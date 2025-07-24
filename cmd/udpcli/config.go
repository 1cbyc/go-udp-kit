package main

import (
	"github.com/spf13/viper"
)

type Config struct {
	Addr    string
	Retries int
	Timeout int
	Backoff float64
	Priority int
	Data    string
}

var cliConfig Config

func loadConfig() {
	viper.SetConfigName("udpcli")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME")
	viper.SetDefault("Addr", "localhost:9090")
	viper.SetDefault("Retries", 3)
	viper.SetDefault("Timeout", 100)
	viper.SetDefault("Backoff", 1.5)
	viper.SetDefault("Priority", 0)
	viper.SetDefault("Data", "hello")
	_ = viper.ReadInConfig()
	cliConfig.Addr = viper.GetString("Addr")
	cliConfig.Retries = viper.GetInt("Retries")
	cliConfig.Timeout = viper.GetInt("Timeout")
	cliConfig.Backoff = viper.GetFloat64("Backoff")
	cliConfig.Priority = viper.GetInt("Priority")
	cliConfig.Data = viper.GetString("Data")
} 