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
		MaxProjectsPerUser   uint8 `mapstructure:"MAX_PROJECTS_PER_USER"`
		MaxMembersPerProject uint8 `mapstructure:"MAX_MEMBERS_PER_PROJECT"`
	} `mapstructure:",squash"`

	Redis struct {
		Password string `mapstructure:"REDIS_PASSWORD"`
		Host     string `mapstructure:"REDIS_HOST"`
		Port     int    `mapstructure:"REDIS_PORT"`
		Database int    `mapstructure:"REDIS_DATABASE"`
	} `mapstructure:",squash"`

	Services struct {
		KafkaBrokers   string `mapstructure:"KAFKA_BROKERS"`
		KafkaUsername  string `mapstructure:"KAFKA_USERNAME"`
		KafkaPassword  string `mapstructure:"KAFKA_PASSWORD"`
		KafkaMechanism string `mapstructure:"KAFKA_MECHANISM"`
		User           string `mapstructure:"USER_SERVICE"`
	} `mapstructure:",squash"`
}
