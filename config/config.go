package config

import (
	"fmt"
	"time"

	appConfig "gitlab.services.mts.ru/salsa/go-base/application/config"
)

type Config struct {
	appConfig.AppConfig
	HTTP                      HTTPConfig            `env:",prefix=HTTP_"`
	MnpEventKafka             appConfig.KafkaConfig `env:",prefix=MNP_EVENT_"`
	PortInOrdersDB            PostgresConfig        `env:",prefix=MNPPORTIN_ORDERS_PG_" validate:"required"`
	PortInCancelDB            PostgresConfig        `env:",prefix=MNPPORTIN_CANCEL_PG_" validate:"required"`
	CDBMessagingDB            PostgresConfig        `env:",prefix=CDB_MESSAGING_PG_" validate:"required"`
	TargetDB                  PostgresConfig        `env:",prefix=MNP_DATAMART_PG_" validate:"required"`
	PortInJobInterval         time.Duration         `env:"PORTIN_JOB_INTERVAL,default=1h"`
	CDBMessageJobInterval     time.Duration         `env:"CDB_MESSAGE_JOB_INTERVAL,default=1h"`
	LookbackDuration          time.Duration         `env:"LOOKBACK_DURATION,default=5m"`
	BatchSize                 int                   `env:"BATCH_SIZE,default=5000"`
	MnpRPSMax                 int                   `env:"MNP_RPS_MAX,default=10"`
	MnpRequestsIntervalMaxSec int                   `env:"MNP_REQUESTS_INTERVAL_MAX_IN_SEC,default=60"`
	MnpRequestEventsLimit     int                   `env:"MNP_REQUEST_EVENTS_LIMIT,default=5"`
	MnpRetryCountMax          int                   `env:"MNP_RETRY_COUNT_MAX,default=10"`
	KafkaEnabled              bool                  `env:"KAFKA_ENABLED,default=false"`
	KafkaTopic                string                `env:"KAFKA_TOPIC"`
	KafkaBootstrap            string                `env:"KAFKA_BOOTSTRAP"`
	KafkaOAuthTokenURL        string                `env:"KAFKA_OAUTH_TOKEN_URL,default=https://isso.mts.ru/auth/realms/mts/protocol/openid-connect/token"`
	KafkaClientID             string                `env:"KAFKA_CLIENT_ID"`
	KafkaClientSecret         string                `env:"KAFKA_CLIENT_SECRET"`
	PortInPrefix              string                `env:"PORTIN_PREFIX,default=pin"`
	PortInCancelTable         string                `env:"PORTIN_CANCEL_TABLE,default=orders"`
	MigrationsPath            string                `env:"MIGRATIONS_PATH,default=/app/db/migrations"`
	MigrationsVersionTable    string                `env:"MIGRATIONS_VERSION_TABLE" validate:"required"`
}

type PostgresConfig struct {
	Host                 string `env:"HOST" validate:"required"`
	Port                 string `env:"PORT,default=5432" validate:"required"`
	DBName               string `env:"DB_NAME" validate:"required"`
	Schema               string `env:"SCHEMA"`
	MigrationUsername    string `env:"MIGRATION_USERNAME" validate:"required"`
	MigrationPassword    string `env:"MIGRATION_PASSWORD" validate:"required"`
	Username             string `env:"USERNAME" validate:"required"`
	Password             string `env:"PASSWORD" validate:"required"`
	SSLMode              string `env:"SSL_MODE,default=disable" validate:"required"`
	MaxConnectionRetries int    `env:"MAX_CONNECTION_RETRIES,default=10" validate:"omitempty"`
}

func (c *PostgresConfig) GetAppConnectionString() string {
	dbConn := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password='%s' sslmode=%s",
		c.Host, c.Port, c.DBName, c.Username, c.Password, c.SSLMode)
	if c.Schema != "" {
		dbConn += fmt.Sprintf(" search_path=%s,public", c.Schema)
	}

	return dbConn
}

func (c *PostgresConfig) GetMigrationConnectionString() string {
	dbConn := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password='%s' sslmode=%s",
		c.Host, c.Port, c.DBName, c.MigrationUsername, c.MigrationPassword, c.SSLMode)
	if c.Schema != "" {
		dbConn += fmt.Sprintf(" search_path=%s,public", c.Schema)
	}

	return dbConn
}

type HTTPConfig struct {
	Port string `env:"PORT" validate:"required"`
}
