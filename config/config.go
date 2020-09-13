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
	BaseUrl   string
	ApiKey    string
	CloudName string
}

func NewConfiguration() *Configuration {
	viper.AutomaticEnv()

	viper.SetDefault("HOST", "0.0.0.0")
	viper.SetDefault("PORT", "4040")
	viper.SetDefault("CDN_BASE_URL", "https://api.cloudinary.com/v1_1")
	viper.SetDefault("CloudName", "sayze")

	return &Configuration{&Http{
		Host: viper.GetString("HOST"),
		Port: viper.GetString("PORT"),
	}, &Cdn{
		BaseUrl:   viper.GetString("CDN_BASE_URL"),
		ApiKey:    viper.GetString("API_KEY"),
		CloudName: viper.GetString("API_KEY"),
	}}

}
