package config

import "github.com/spf13/viper"

type Configuration struct {
	Http *Http
	Cdn  *Cdn
}

type Http struct {
	Host string
	Port string
}

type Cdn struct {
	Host string
	Port string
	Base string
}

func NewConfiguration() *Configuration {
	viper.AutomaticEnv()

	viper.SetDefault("HTTP_HOST", "localhost")
	viper.SetDefault("HTTP_PORT", "4040")
	viper.SetDefault("CDN_HOST", "localhost")
	viper.SetDefault("CDN_PORT", "6060")
	viper.SetDefault("CDN_BASE", "friendly")

	return &Configuration{&Http{
		Host: viper.GetString("HTTP_HOST"),
		Port: viper.GetString("HTTP_PORT"),
	}, &Cdn{
		Host: viper.GetString("CDN_HOST"),
		Port: viper.GetString("CDN_PORT"),
		Base: viper.GetString("CDN_BASE"),
	}}

}
