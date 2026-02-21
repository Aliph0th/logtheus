package config

type AppConfig struct {
	Env       string `mapstructure:"ENV"`
	AppDomain string `mapstructure:"APP_DOMAIN"`
	Server    struct {
		Port int `mapstructure:"PORT"`
	} `mapstructure:",squash"`
	SMTP struct {
		Host     string `mapstructure:"MAIL_HOST"`
		Port     int    `mapstructure:"MAIL_PORT"`
		Username string `mapstructure:"MAIL_LOGIN"`
		Password string `mapstructure:"MAIL_PASSWORD"`
		From     string `mapstructure:"MAIL_FROM"`
	} `mapstructure:",squash"`
}
