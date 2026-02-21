package config

type AppConfig struct {
	Env    string `mapstructure:"ENV"`
	Server struct {
		Port int `mapstructure:"PORT"`
	} `mapstructure:",squash"`
	DB struct {
		Host     string `mapstructure:"POSTGRES_HOST"`
		Port     int    `mapstructure:"POSTGRES_PORT"`
		Name     string `mapstructure:"POSTGRES_DB"`
		User     string `mapstructure:"POSTGRES_USER"`
		Password string `mapstructure:"POSTGRES_PASSWORD"`
	} `mapstructure:",squash"`

	JWT struct {
		AccessSecret  string `mapstructure:"JWT_ACCESS_SECRET"`
		RefreshSecret string `mapstructure:"JWT_REFRESH_SECRET"`
		Issuer        string `mapstructure:"JWT_ISSUER"`
	} `mapstructure:",squash"`

	Redis struct {
		Password string `mapstructure:"REDIS_PASSWORD"`
		Host     string `mapstructure:"REDIS_HOST"`
		Port     int    `mapstructure:"REDIS_PORT"`
		Database int    `mapstructure:"REDIS_DATABASE"`
	} `mapstructure:",squash"`

	Services struct {
		Mail string `mapstructure:"MAIL_SERVICE"`
	} `mapstructure:",squash"`
}
