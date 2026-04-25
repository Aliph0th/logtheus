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

	Postgres struct {
		Host     string `mapstructure:"POSTGRES_HOST"`
		Port     int    `mapstructure:"POSTGRES_PORT"`
		Name     string `mapstructure:"POSTGRES_DB"`
		User     string `mapstructure:"POSTGRES_USER"`
		Password string `mapstructure:"POSTGRES_PASSWORD"`
	} `mapstructure:",squash"`

	S3 struct {
		AccessKey string `mapstructure:"S3_ACCESS_KEY"`
		SecretKey string `mapstructure:"S3_SECRET_KEY"`
		Host      string `mapstructure:"S3_HOST"`
		Region    string `mapstructure:"S3_REGION"`
	} `mapstructure:",squash"`

	Kafka struct {
		Brokers                          string `mapstructure:"KAFKA_BROKERS"`
		Username                         string `mapstructure:"KAFKA_USERNAME"`
		Password                         string `mapstructure:"KAFKA_PASSWORD"`
		Mechanism                        string `mapstructure:"KAFKA_MECHANISM"`
		LogsConsumerBatchMaxMessages     int    `mapstructure:"LOGS_CONSUMER_BATCH_MAX_MESSAGES"`
		LogsConsumerBatchMaxBytes        int    `mapstructure:"LOGS_CONSUMER_BATCH_MAX_BYTES"`
		LogsConsumerBatchMaxWaitMs       int    `mapstructure:"LOGS_CONSUMER_BATCH_MAX_WAIT_MS"`
		FeaturesConsumerBatchMaxMessages int    `mapstructure:"FEATURES_CONSUMER_BATCH_MAX_MESSAGES"`
		FeaturesConsumerBatchMaxBytes    int    `mapstructure:"FEATURES_CONSUMER_BATCH_MAX_BYTES"`
		FeaturesConsumerBatchMaxWaitMs   int    `mapstructure:"FEATURES_CONSUMER_BATCH_MAX_WAIT_MS"`
	} `mapstructure:",squash"`

	Persistence struct {
		LogsCHInsertBatchSize int `mapstructure:"LOGS_CH_INSERT_BATCH_SIZE"`
		LogsDedupeTTLSeconds  int `mapstructure:"LOGS_DEDUPE_TTL_SECONDS"`
		LogsDedupeCacheSize   int `mapstructure:"LOGS_DEDUPE_CACHE_SIZE"`
	} `mapstructure:",squash"`

	Services struct {
		Project string `mapstructure:"PROJECT_SERVICE"`
	} `mapstructure:",squash"`
}
