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

	Settings struct {
		MaxApplicationsPerProject uint8 `mapstructure:"MAX_APPLICATIONS_PER_PROJECT"`
	} `mapstructure:",squash"`

	ApiKey struct {
		BytesLength int    `mapstructure:"API_KEY_BYTES"`
		Secret      string `mapstructure:"API_KEY_SECRET"`
	} `mapstructure:",squash"`

	Services struct {
		Mail    string `mapstructure:"MAIL_SERVICE"`
		User    string `mapstructure:"USER_SERVICE"`
		Project string `mapstructure:"PROJECT_SERVICE"`
	} `mapstructure:",squash"`
}
