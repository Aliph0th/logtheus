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

	Services struct {
		Application string `mapstructure:"APPLICATION_SERVICE"`
		Project     string `mapstructure:"PROJECT_SERVICE"`
		LogEngine   string `mapstructure:"LOG_ENGINE_SERVICE"`
	} `mapstructure:",squash"`
}
