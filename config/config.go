package config

import "github.com/spf13/viper"

type Configuration struct {
	Http *Http
}

type Http struct {
	Host string
	Port string
}

func NewConfiguration() *Configuration {
	viper.AutomaticEnv()

	viper.SetDefault("HOST", "localhost")
	viper.SetDefault("PORT", "4040")

	return &Configuration{&Http{
		Host: viper.GetString("HOST"),
		Port: viper.GetString("PORT"),
	}}

}