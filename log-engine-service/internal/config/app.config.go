package config

type AppConfig struct {
	Env    string `mapstructure:"ENV"`
	Server struct {
		Port int `mapstructure:"PORT"`
	} `mapstructure:",squash"`
	ClickHouse struct {
		Host     string `mapstructure:"CLICKHOUSE_HOST"`
		Port     int    `mapstructure:"CLICKHOUSE_PORT"`
		Name     string `mapstructure:"CLICKHOUSE_DB"`
		User     string `mapstructure:"CLICKHOUSE_USER"`
		Password string `mapstructure:"CLICKHOUSE_PASSWORD"`
	} `mapstructure:",squash"`

	S3 struct {
		AccessKey string `mapstructure:"S3_ACCESS_KEY"`
		SecretKey string `mapstructure:"S3_SECRET_KEY"`
		Host      string `mapstructure:"S3_HOST"`
		Region    string `mapstructure:"S3_REGION"`
	} `mapstructure:",squash"`

	Kafka struct {
		Brokers   string `mapstructure:"KAFKA_BROKERS"`
		Username  string `mapstructure:"KAFKA_USERNAME"`
		Password  string `mapstructure:"KAFKA_PASSWORD"`
		Mechanism string `mapstructure:"KAFKA_MECHANISM"`
	} `mapstructure:",squash"`

	Services struct{} `mapstructure:",squash"`
}
