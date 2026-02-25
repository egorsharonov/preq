package config

import (
	"fmt"

	appConfig "gitlab.services.mts.ru/salsa/go-base/application/config"
	"gitlab.services.mts.ru/salsa/go-base/application/validators"

	"gitlab.services.mts.ru/salsa/mnp-hub/portin-requests/internal/model"
)

type Config struct {
	appConfig.AppConfig
	HTTP                   HTTPConfig            `env:",prefix=HTTP_"`
	MnpEventKafka          appConfig.KafkaConfig `env:",prefix=MNP_EVENT_"`
	PortInDB               PostgresConfig        `env:",prefix=MNPPORTIN_PG_" validate:"required"`
	MigrationsPath         string                `env:"MIGRATIONS_PATH,default=/app/db/migrations"`
	MigrationsVersionTable string                `env:"MIGRATIONS_VERSION_TABLE" validate:"required"`
	FTP                    appConfig.FTPConfig   `env:",prefix=FTP_"`
	MTS                    MTSConfig             `env:",prefix=MTS_"`
	PortInOrders           PortInOrdersConfig    `env:",prefix=PORTIN_ORDERS_"`
	Consul                 ConsulConfig          `env:",prefix=CONSUL_"`
}

func (cfg *Config) ApplyDefaults() {
	cfg.PortInOrders.ApplyDefaults()
}

func (cfg *Config) Validate() error {
	return validators.ValidateStruct(cfg)
}

type PortInOrdersConfig struct {
	// DisableOpenOrdersCheck - отключить проверку открытых заявок. Загружается только из Consul.
	DisableOpenOrdersCheck bool
	// AllowedStatusesForNewOrder - список статусов, при которых разрешено создание новой заявки. Загружается только из Consul.
	AllowedStatusesForNewOrder []int
	// PortationNumbersStatesArgegateMap - Дополнительно для статусов завершения малой портации,
	// т.к. для каждого номера сообщения приходят отдельно, нужна агрегация c вычислением общего статуса заявки.
	//
	// Мап статусов переносимых номеров в множество статусов, в котором должны лежать все номера в заявке,
	// для применения указанного статуса ко всей заявке и соотвествующий агренированный статус заявки
	PortationNumbersStatesAgregateMap map[string]model.PortationNumberStateAgregate
}

func (cfg *PortInOrdersConfig) ApplyDefaults() {
	if len(cfg.AllowedStatusesForNewOrder) == 0 {
		cfg.AllowedStatusesForNewOrder = model.GetDefaultAllowedStatusesForNewOrder()
	}

	if len(cfg.PortationNumbersStatesAgregateMap) == 0 {
		cfg.PortationNumbersStatesAgregateMap = model.GetDefaultPortationNumbersStatesAgregateMap()
	}
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

type MTSConfig struct {
	RecipientRN      string `env:"RECIPIENT_RN" validate:"required"`
	RecipientName    string `env:"RECIPIENT_NAME" validate:"required"`
	RecipientCdbCode string `env:"RECIPIENT_CDBCODE" validate:"required"`
}

type ConsulConfig struct {
	Address string `env:"HTTP_ADDR" validate:"required,http_url"`
	Token   string `env:"HTTP_TOKEN"`
	Key     string `env:"KEY" validate:"required"`
}
