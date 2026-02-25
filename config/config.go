package config

import (
	"fmt"

	appConfig "gitlab.services.mts.ru/salsa/go-base/application/config"
)

type Config struct {
	appConfig.AppConfig
	HTTP                   HTTPConfig            `env:",prefix=HTTP_"`
	MnpEventKafka          appConfig.KafkaConfig `env:",prefix=MNP_EVENT_"`
	PortInDB               PostgresConfig        `env:",prefix=MNPPORTIN_PG_" validate:"required"`
	MigrationsPath         string                `env:"MIGRATIONS_PATH,default=/app/db/migrations"`
	MigrationsVersionTable string                `env:"MIGRATIONS_VERSION_TABLE" validate:"required"`
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
