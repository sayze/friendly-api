package config

import "github.com/spf13/viper"

type Configuration struct {
	Http *Http
	Cdn  *Cdn
}

type Http struct {
	Port string
}

type Cdn struct {
	UploadUrl string
	ImageUrl  string
	ApiKey    string
	ApiSecret string
}

func NewConfiguration() *Configuration {
	viper.AutomaticEnv()
	viper.SetDefault("PORT", "4040")
	viper.SetDefault("CDN_UPLOAD_URL", "https://api.cloudinary.com/v1_1/sayze/image/upload")
	viper.SetDefault("CDN_IMAGE_URL", "https://res.cloudinary.com/sayze/image/upload")

	return &Configuration{&Http{
		Port: viper.GetString("PORT"),
	}, &Cdn{
		UploadUrl: viper.GetString("CDN_UPLOAD_URL"),
		ImageUrl:  viper.GetString("CDN_IMAGE_URL"),
		ApiKey:    viper.GetString("CDN_API_KEY"),
		ApiSecret: viper.GetString("CDN_API_SECRET"),
	}}

}
