package config

type AppConfig struct {
	Env    string `mapstructure:"ENV"`
	Server struct {
		Port int `mapstructure:"PORT"`
	} `mapstructure:",squash"`

	Services struct {
		User    string `mapstructure:"USER_SERVICE"`
		Project string `mapstructure:"PROJECT_SERVICE"`
	} `mapstructure:",squash"`
}
