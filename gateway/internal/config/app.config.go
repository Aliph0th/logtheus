package config

type AppConfig struct {
	Env    string `mapstructure:"ENV"`
	Server struct {
		Port        int    `mapstructure:"PORT"`
		AllowOrigin string `mapstructure:"ALLOWED_ORIGIN"`
	} `mapstructure:",squash"`

	Services struct {
		User        string `mapstructure:"USER_SERVICE"`
		Project     string `mapstructure:"PROJECT_SERVICE"`
		Application string `mapstructure:"APPLICATION_SERVICE"`
		Ingestion   string `mapstructure:"INGESTION_SERVICE"`
		LogEngine   string `mapstructure:"LOG_ENGINE_SERVICE"`
	} `mapstructure:",squash"`
}
